package core

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type ServerConfig struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

type HookConfig struct {
	Plugin string `json:"plugin"`
}

type Config struct {
	Server ServerConfig          `json:"server"`
	Hooks  map[string]HookConfig `json:"hooks"`
}

var DefaultConfig = Config{
	Server: ServerConfig{
		Hostname: "0.0.0.0",
		Port:     3000,
	},
	Hooks: map[string]HookConfig{},
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig, err
	}
	cfg := DefaultConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig, err
	}
	return cfg, nil
}
