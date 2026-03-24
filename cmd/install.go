package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rajpootathar/megahorn/internal/config"
	"github.com/spf13/cobra"
)

func writeMCPConfig(settingsPath, binPath string) error {
	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]any)
	}

	mcpServers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
	}

	mcpServers["megahorn"] = map[string]any{
		"command": binPath,
		"args":    []string{"server"},
	}
	settings["mcpServers"] = mcpServers

	os.MkdirAll(filepath.Dir(settingsPath), 0755)

	out, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(settingsPath, out, 0644)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Configure megahorn MCP server in Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.EnsureDir()

		home, _ := os.UserHomeDir()
		settingsPath := filepath.Join(home, ".claude", "settings.json")

		binPath, err := os.Executable()
		if err != nil {
			binPath = "megahorn"
		}

		if err := writeMCPConfig(settingsPath, binPath); err != nil {
			return fmt.Errorf("failed to write MCP config: %w", err)
		}

		fmt.Printf("MCP server configured in %s\n", settingsPath)
		fmt.Println("Restart Claude Code to activate megahorn tools.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
