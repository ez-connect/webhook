package plugin

import (
	"errors"
	"net/http"
	"plugin"
)

type (
	NewFn = func() Plugin

	Plugin interface {
		Start(config string) error
		Stop() error
		Get(req *http.Request) ([]byte, error)
		Post(req *http.Request) ([]byte, error)
		Delete(req *http.Request) ([]byte, error)
	}
)

func LoadPlugin(so string) (Plugin, error) {
	var (
		p     *plugin.Plugin
		newFn plugin.Symbol
		err   error
	)

	if p, err = plugin.Open(so); err != nil {
		return nil, err
	}

	if newFn, err = p.Lookup("New"); err != nil {
		return nil, err
	}

	if v, ok := newFn.(NewFn); ok {
		return v(), nil
	}

	return nil, errors.New("plugin not implemented `func New() Plugin`")
}
