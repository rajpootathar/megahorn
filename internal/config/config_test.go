package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Browser.Headed != false {
		t.Error("expected headed=false by default")
	}
	if len(cfg.Platforms.Defaults) != 0 {
		t.Error("expected empty defaults when no config file")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := []byte("browser:\n  headed: true\nplatforms:\n  defaults:\n    - twitter\n    - linkedin\n")
	os.WriteFile(cfgPath, content, 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Browser.Headed != true {
		t.Error("expected headed=true from config file")
	}
	if len(cfg.Platforms.Defaults) != 2 {
		t.Errorf("expected 2 defaults, got %d", len(cfg.Platforms.Defaults))
	}
}

func TestConfigDir(t *testing.T) {
	dir := Dir()
	if dir == "" {
		t.Error("config dir should not be empty")
	}
}

func TestSetValue(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := []byte("browser:\n  headed: false\n")
	os.WriteFile(cfgPath, content, 0644)

	err := SetValue(cfgPath, "browser.headed", "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, _ := Load(cfgPath)
	if cfg.Browser.Headed != true {
		t.Error("expected headed=true after set")
	}
}
