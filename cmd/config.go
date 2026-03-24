package cmd

import (
	"fmt"

	"github.com/rajpootathar/megahorn/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		fmt.Printf("Config directory:   %s\n", config.Dir())
		fmt.Printf("Browser headed:     %v\n", cfg.Browser.Headed)
		fmt.Printf("Chrome path:        %s\n", cfg.Browser.ChromePath)
		fmt.Printf("Default platforms:  %v\n", cfg.Platforms.Defaults)
		fmt.Printf("Default subreddits: %v\n", cfg.Reddit.DefaultSubreddits)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.EnsureDir()
		if err := config.SetValue(config.DefaultPath(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}
		fmt.Printf("Set %s = %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
