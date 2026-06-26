package config

import (
	"os"

	"edge-app/internal/connector"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		HTTPPort int `yaml:"http_port"`
	} `yaml:"server"`

	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`

	Aggregator struct {
		WindowMs int64 `yaml:"window_ms"`
	} `yaml:"aggregator"`

	Connectors connector.Config `yaml:"connectors"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
