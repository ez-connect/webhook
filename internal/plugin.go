package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/ez-connect/webhook/pkg/core"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type PluginManager struct {
	runtime wazero.Runtime
	hooks   map[string]wazero.CompiledModule
}

func NewPluginManager(ctx context.Context) *PluginManager {
	r := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	return &PluginManager{
		runtime: r,
		hooks:   make(map[string]wazero.CompiledModule),
	}
}

func (pm *PluginManager) RegisterPlugin(ctx context.Context, hookPath string, pluginPath string) error {
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return fmt.Errorf("read plugin file %q: %w", pluginPath, err)
	}

	compiled, err := pm.runtime.CompileModule(ctx, data)
	if err != nil {
		return fmt.Errorf("compile plugin %q: %w", pluginPath, err)
	}

	pm.hooks[hookPath] = compiled
	return nil
}

func (pm *PluginManager) RunPlugin(ctx context.Context, hookPath string, req *http.Request) (*core.PluginResponse, error) {
	compiled, ok := pm.hooks[hookPath]
	if !ok {
		return nil, fmt.Errorf("no plugin registered for hook %q", hookPath)
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	var bodyJSON any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &bodyJSON); err != nil {
			bodyJSON = string(body)
		}
	}

	// Extract headers
	headers := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Create plugin request
	data := core.PluginRequest{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: headers,
		Body:    bodyJSON,
	}

	// Marshal request to JSON
	requestJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal plugin request: %w", err)
	}

	// Prepare stdin, stdout, and stderr
	stdin := bytes.NewReader(requestJSON)
	stdout := bytes.Buffer{}
	var stderr bytes.Buffer

	// Initialize the plugin with stdin/stdout/stderr
	moduleConfig := wazero.NewModuleConfig().
		WithStartFunctions().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(&stderr)

	module, err := pm.runtime.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		return nil, fmt.Errorf("instantiate plugin: %w", err)
	}
	defer module.Close(ctx)

	// Get and call the run function
	runFunc := module.ExportedFunction("run")
	if runFunc == nil {
		return nil, fmt.Errorf("plugin does not export a run function")
	}

	// Call the run function (no parameters needed)
	_, err = runFunc.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("plugin run failed: %w", err)
	}

	// Print stderr if something was written there, useful for debugging wasm panics
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "Plugin stderr (%s): %s\n", hookPath, stderr.String())
	}

	// Parse response from stdout
	var pluginRes core.PluginResponse
	outBytes := stdout.Bytes()
	if len(outBytes) == 0 {
		return nil, fmt.Errorf("plugin returned no output")
	}

	if err := json.Unmarshal(outBytes, &pluginRes); err != nil {
		return nil, fmt.Errorf("unmarshal plugin response: %w (output: %s)", err, string(outBytes))
	}

	// Default status to 200 if not set
	if pluginRes.Status == 0 {
		pluginRes.Status = 200
	}

	// Call destroy if it exists
	destroyFunc := module.ExportedFunction("destroy")
	if destroyFunc != nil {
		if _, err := destroyFunc.Call(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Plugin destroy failed: %v\n", err)
		}
	}

	return &pluginRes, nil
}

// ExecCommandFromWasm executes a host command from WASM module.
// Supports commands like: "podman exec <name> ls", "docker exec ...", etc.
// Returns: output string and exit code
func (pm *PluginManager) ExecCommandFromWasm(cmdStr string) (string, int32) {
	// Whitelist of allowed root commands for security
	allowedCommands := map[string]bool{
		"podman": true,
		"docker": true,
		"ls":     true,
		"cat":    true,
		"pwd":    true,
		"echo":   true,
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "error: empty command", 1
	}

	cmd := parts[0]
	if !allowedCommands[cmd] {
		return fmt.Sprintf("error: command not allowed: %s", cmd), 1
	}

	execCmd := exec.Command(cmd, parts[1:]...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()
	exitCode := int32(0)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = int32(exitErr.ExitCode())
		} else {
			exitCode = 1
		}
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output = stderr.String()
	}

	return output, exitCode
}
