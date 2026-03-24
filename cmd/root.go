package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "megahorn",
	Short: "One binary to rule them all — cross-post to social media",
	Long:  "Megahorn posts to Twitter, LinkedIn, and Reddit from your terminal or via MCP for AI agents.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
