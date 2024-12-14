package service

import (
	"fmt"
	"log/slog"
	"net/http"

	"freemind.com/webhook/plugin"
	base "freemind.com/webhook/plugin"
	"freemind.com/webhook/service/health"
	"freemind.com/webhook/service/index"
)

type Hook struct {
	Plugin string `yaml:"plugin"` // path/to/plugin.so
	Config string // path/to/config.yaml
	Secret string // secret key
}

const (
	ServerPort = ":8080" // serve command on this port
)

var (
	opts    map[string]*Hook
	plugins = map[string]plugin.Plugin{}
)

func ServeHTTP(config string) error {
	if err := load(config); err != nil {
		return err
	}

	http.HandleFunc("/", index.Get)
	http.HandleFunc("/health", health.Get)

	return http.ListenAndServe(ServerPort, nil)
}

func load(config string) error {
	slog.Info("load config", "path", config)
	if err := base.ReadConfig(config, &opts); err != nil {
		return err
	}

	for k, v := range opts {
		if err := addHook(k, v.Plugin, v.Config); err != nil {
			slog.Error("add hook", "id", k, "err", err.Error())
		}
	}

	return nil
}

func addHook(id, so, config string) error {
	var (
		p   plugin.Plugin
		err error
	)

	slog.Info("add hook", "id", id, "so", so, "config", config)

	if p, err = plugin.LoadPlugin(so); err != nil {
		return err
	}

	if err = p.Start(config); err != nil {
		return err
	}

	plugins[id] = p

	pattern := fmt.Sprintf("/hooks/%s", id)
	http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		var (
			buf []byte
			err error
		)
		slog.Info(fmt.Sprintf("%s %s", req.Method, pattern), "q", req.URL.Query().Encode())
		switch req.Method {
		case http.MethodGet:
			if buf, err = p.Get(req); err != nil {
				handleError(w, err)
				return
			}

			w.Write(buf)

		case http.MethodPost:
			defer req.Body.Close()

			if buf, err = p.Post(req); err != nil {
				handleError(w, err)
				return
			}

			w.Write(buf)

		case http.MethodDelete:
			http.HandleFunc(pattern, nil)
			w.Write([]byte("TODO: Implement delete a endpoint"))
		}
	})

	return nil
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
