package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var ctx context.Context
var pm *PluginManager

func loadPlugins() {
	ctx = context.Background()
	pm = NewPluginManager(ctx)

	for path, hook := range config.Hooks {
		err := pm.RegisterPlugin(ctx, path, hook.Plugin)
		if err != nil {
			log.Fatalf("Failed to register plugin for %s: %v", path, err)
		}
		log.Printf("Registered %s -> %s\n", path, hook.Plugin)
	}
}

func hook(w http.ResponseWriter, r *http.Request) {
	// Find the hook path, ignoring query strings.
	hookPath := strings.TrimPrefix(r.URL.Path, "/")
	if hookPath == "" || !strings.HasPrefix(r.URL.Path, "/") {
		hookPath = r.URL.Path
	}

	// Map back to the config path if needed
	matchPath := r.URL.Path
	if _, ok := config.Hooks[matchPath]; !ok {
		// fallback to stripping leading slash if not matched?
		// The node app mapped /:hook directly.
		// Let's assume the hook path in config is "/hook-name"
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	log.Printf("POST %s\n", matchPath)

	res, err := pm.RunPlugin(ctx, matchPath, r)
	if err != nil {
		log.Printf("Plugin error: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   "Failed to process hook",
			"message": err.Error(),
		})
		return
	}

	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(res.Status)

	if bodyStr, ok := res.Body.(string); ok && (strings.HasPrefix(bodyStr, "{") || strings.HasPrefix(bodyStr, "[")) && w.Header().Get("Content-Type") == "application/json" {
		// Valid JSON string, but already passed back; since res.Body is `any`,
		// To serialize correctly:
		// If Body is a string, perhaps the plugin serialized it already.
		if err := json.NewEncoder(w).Encode(res.Body); err != nil {
			w.Write([]byte(bodyStr))
		}
	} else {
		json.NewEncoder(w).Encode(res.Body)
	}
}
