package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallWritesMCPConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude.json")

	err := writeMCPConfig(configPath, "/usr/local/bin/megahorn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
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
	if megahorn["type"] != "stdio" {
		t.Errorf("expected type stdio, got: %v", megahorn["type"])
	}
}

func TestInstallPreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude.json")

	// Write existing config
	existing := map[string]any{
		"numStartups": 42,
		"mcpServers": map[string]any{
			"other-server": map[string]any{
				"type":    "stdio",
				"command": "other",
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(configPath, data, 0644)

	err := writeMCPConfig(configPath, "/usr/local/bin/megahorn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := os.ReadFile(configPath)
	var settings map[string]any
	json.Unmarshal(result, &settings)

	// Verify existing data preserved
	if settings["numStartups"] != float64(42) {
		t.Error("existing config was not preserved")
	}

	mcpServers := settings["mcpServers"].(map[string]any)
	if _, ok := mcpServers["other-server"]; !ok {
		t.Error("existing MCP server was removed")
	}
	if _, ok := mcpServers["megahorn"]; !ok {
		t.Error("megahorn was not added")
	}
}
