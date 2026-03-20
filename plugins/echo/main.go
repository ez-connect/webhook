package main

import (
	"encoding/json"
	"io"
	"os"
	"unsafe"

	"github.com/ez-connect/webhook/pkg/core"
)

// // Malloc allocates memory in the WASM module and returns a pointer.
// // This is required for passing data between the host and WASM modules.
// //
// //export allocate
// func Allocate(size uint32) uintptr {
// 	buf := make([]byte, size)
// 	//nolint:gosec // Required for WASM FFI memory management
// 	return uintptr(unsafe.Pointer(&buf[0]))
// }

// // Free frees allocated memory in the WASM module.
// // In Go, memory is automatically managed by the garbage collector.
// //
// //export deallocate
// func Deallocate(_ uintptr) {
// 	// Memory is managed by Go's garbage collector, this is a no-op
// }

// Init initializes the plugin when it's loaded.
//
//export init
func Init() {
	// Plugin initialization can be added here
}

// Destroy cleans up the plugin resources when it's unloaded.
//
//export destroy
func Destroy() {
	// Plugin cleanup can be added here
}

// Run processes the HTTP request and returns a response.
// The request data is read from stdin as JSON, and the response is written to stdout as JSON.
//
//export run
func Run() {
	// Read the request from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		writeError(500, "Failed to read request from stdin: "+err.Error())
		return
	}

	var req core.PluginRequest
	if err := json.Unmarshal(data, &req); err != nil {
		writeError(500, "Failed to parse request: "+err.Error())
		return
	}

	// Echo response with request data
	res := core.PluginResponse{
		Status: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]any{
			"url":     req.URL,
			"method":  req.Method,
			"headers": req.Headers,
			"body":    req.Body,
		},
	}

	out, err := json.Marshal(res)
	if err != nil {
		writeError(500, "Failed to marshal response: "+err.Error())
		return
	}

	// Write the response to stdout (captured by plugin manager)
	if _, err := os.Stdout.Write(out); err != nil {
		os.Stderr.WriteString("Failed to write response: " + err.Error() + "\n")
	}
}

// ExecCommand allows the WASM module to request command execution on the host.
// This function is exported for future use when host command execution is implemented.
// Parameters:
//   - cmdPtr: pointer to the command string in WASM memory
//   - cmdSize: size of the command string
//   - unused: reserved for output pointer (future implementation)
//   - unused: reserved for output size (future implementation)
//
// Returns: exit code (0 for success, non-zero for error)
//
//export exec_command
func ExecCommand(cmdPtr uint32, cmdSize uint32, _ uint32, _ uint32) uint32 {
	// This function provides a hook for WASM to execute commands on the host.
	// The actual implementation is delegated to the host runtime.
	// This is necessary for WASM modules to execute system commands.
	//nolint:gosec // WASM FFI: pointer is controlled by trusted host
	cmdData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(cmdPtr))), cmdSize)
	os.Stderr.WriteString("Command execution requested: " + string(cmdData) + "\n")
	return 1 // Return error code - not yet implemented in WASM
}

// writeError writes an error response to stdout.
// The response is formatted as JSON with the given status code and message.
func writeError(status int, message string) {
	res := PluginResponse{
		Status: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]string{
			"error":   "error",
			"message": message,
		},
	}

	out, _ := json.Marshal(res)
	_, _ = os.Stdout.Write(out)
}

// main is the entry point for the Go program.
// In WASM execution environment, this is not called.
func main() {
}
