package cmd

import (
	"context"
	"fmt"
	"os"

	mcpserver "github.com/rajpootathar/megahorn/internal/mcp"
	"github.com/rajpootathar/megahorn/internal/auth"
	"github.com/rajpootathar/megahorn/internal/config"
	"github.com/rajpootathar/megahorn/internal/platform"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start MCP server (stdio JSON-RPC)",
	Long:  "Start megahorn as an MCP server for AI agents like Claude Code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(config.DefaultPath())
		kr := auth.NewKeyring()

		registry := platform.NewRegistry()
		registry.Register(platform.NewTwitter(kr, cfg))
		registry.Register(platform.NewLinkedIn(kr, cfg))
		registry.Register(platform.NewReddit(kr, cfg))

		mcpSrv := mcpserver.NewServer(registry)
		stdioSrv := server.NewStdioServer(mcpSrv)
		fmt.Fprintln(os.Stderr, "megahorn MCP server started (stdio)")
		return stdioSrv.Listen(context.Background(), os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
