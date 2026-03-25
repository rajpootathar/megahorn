package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rajpootathar/megahorn/internal/config"
	"github.com/spf13/cobra"
)

// installViaCLI tries to register via `claude mcp add` (the official way).
// Returns true if it succeeded.
func installViaCLI(binPath string) bool {
	claude, err := exec.LookPath("claude")
	if err != nil {
		return false
	}
	cmd := exec.Command(claude, "mcp", "add", "megahorn", "-s", "user", "--", binPath, "server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

// writeMCPConfig writes the MCP server entry directly to ~/.claude.json
// (the file Claude Code actually reads MCP servers from).
func writeMCPConfig(configPath, binPath string) error {
	var settings map[string]any
	data, err := os.ReadFile(configPath)
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
		"type":    "stdio",
		"command": binPath,
		"args":    []string{"server"},
	}
	settings["mcpServers"] = mcpServers

	os.MkdirAll(filepath.Dir(configPath), 0755)

	out, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(configPath, out, 0644)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Configure megahorn MCP server in Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.EnsureDir()

		binPath, err := os.Executable()
		if err != nil {
			binPath = "megahorn"
		}

		// Try the official CLI first
		if installViaCLI(binPath) {
			fmt.Println("Restart Claude Code to activate megahorn tools.")
			return nil
		}

		// Fallback: write directly to ~/.claude.json
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".claude.json")

		if err := writeMCPConfig(configPath, binPath); err != nil {
			return fmt.Errorf("failed to write MCP config: %w", err)
		}

		fmt.Printf("MCP server configured in %s\n", configPath)
		fmt.Println("Restart Claude Code to activate megahorn tools.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
