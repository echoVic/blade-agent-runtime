package config

import (
	"os"

	"github.com/spf13/viper"
)

type Manager struct {
	Path string
}

func NewManager(path string) *Manager {
	return &Manager{Path: path}
}

func (m *Manager) Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(m.Path)
	if _, err := os.Stat(m.Path); err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (m *Manager) Save(cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(m.Path)
	v.Set("version", cfg.Version)
	v.Set("git", cfg.Git)
	v.Set("policy", cfg.Policy)
	v.Set("hooks", cfg.Hooks)
	v.Set("output", cfg.Output)
	return v.WriteConfigAs(m.Path)
}
