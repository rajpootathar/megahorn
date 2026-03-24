package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/rajpootathar/megahorn/internal/platform"
)

func NewServer(registry *platform.Registry) *server.MCPServer {
	s := server.NewMCPServer(
		"megahorn",
		"0.1.0",
	)

	RegisterTools(s, registry)

	return s
}
