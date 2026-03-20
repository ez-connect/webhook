package internal

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ez-connect/webhook/pkg/core"
)

var config core.Config

func Serve(c core.Config) {
	config = c

	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboard)
	mux.HandleFunc("GET /health", health)

	loadPlugins()
	mux.HandleFunc("POST /{hook}", hook)

	addr := fmt.Sprintf("%s:%d", config.Server.Hostname, config.Server.Port)
	log.Printf("Webhook server running on http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
