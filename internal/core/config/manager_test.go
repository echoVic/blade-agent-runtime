package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_LoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	cfg := DefaultConfig()
	cfg.Git.DefaultBase = "develop"
	cfg.Policy.Enabled = true

	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != 1 {
		t.Errorf("expected Version 1, got %d", loaded.Version)
	}
	if loaded.Git.DefaultBase != "develop" {
		t.Errorf("expected DefaultBase 'develop', got '%s'", loaded.Git.DefaultBase)
	}
	if !loaded.Policy.Enabled {
		t.Error("expected Policy.Enabled to be true")
	}
}

func TestManager_LoadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	m := NewManager(configPath)

	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("expected default config with Version 1, got %d", cfg.Version)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != 1 {
		t.Errorf("expected Version 1, got %d", cfg.Version)
	}
	if cfg.Git.DefaultBase != "main" {
		t.Errorf("expected DefaultBase 'main', got '%s'", cfg.Git.DefaultBase)
	}
	if cfg.Git.BranchPrefix != "bar/" {
		t.Errorf("expected BranchPrefix 'bar/', got '%s'", cfg.Git.BranchPrefix)
	}
	if cfg.Policy.Enabled {
		t.Error("expected Policy.Enabled to be false by default")
	}
	if !cfg.Output.Color {
		t.Error("expected Output.Color to be true by default")
	}
}

func TestManager_SaveCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	cfg := DefaultConfig()
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}
