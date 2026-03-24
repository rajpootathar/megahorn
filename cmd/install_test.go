package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallWritesMCPConfig(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")

	err := writeMCPConfig(settingsPath, "/usr/local/bin/megahorn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	mcpServers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("expected mcpServers key")
	}
	megahorn, ok := mcpServers["megahorn"].(map[string]any)
	if !ok {
		t.Fatal("expected megahorn key")
	}
	if megahorn["command"] != "/usr/local/bin/megahorn" {
		t.Errorf("unexpected command: %v", megahorn["command"])
	}
}
