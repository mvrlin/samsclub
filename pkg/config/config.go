package config

import (
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents a configuration file.
type Config struct {
}

// New creates an instance of Config.
func New() (*Config, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	cfgPath := path.Join(dir, "config.toml")

	if _, err = toml.DecodeFile(cfgPath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
