package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rajpootathar/megahorn/internal/platform"
)

type ToolDef struct {
	Name        string
	Description string
}

func BuildToolList() []ToolDef {
	return []ToolDef{
		{Name: "megahorn_post", Description: "Post content to a social media platform"},
		{Name: "megahorn_auth_status", Description: "Check which platforms are authenticated"},
		{Name: "megahorn_auth", Description: "Initiate authentication for a platform (opens browser)"},
	}
}

func RegisterTools(s *server.MCPServer, registry *platform.Registry) {
	// megahorn_post
	s.AddTool(mcplib.Tool{
		Name:        "megahorn_post",
		Description: "Post content to a social media platform",
		InputSchema: mcplib.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"platform":  map[string]any{"type": "string", "enum": []string{"twitter", "linkedin", "reddit"}},
				"content":   map[string]any{"type": "string", "description": "The post content"},
				"subreddit": map[string]any{"type": "string", "description": "Target subreddit(s), comma-separated (reddit only)"},
				"dry_run":   map[string]any{"type": "boolean", "default": false},
			},
			Required: []string{"platform", "content"},
		},
	}, handlePost(registry))

	// megahorn_auth_status
	s.AddTool(mcplib.Tool{
		Name:        "megahorn_auth_status",
		Description: "Check which platforms are authenticated and their token expiry status",
		InputSchema: mcplib.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, handleAuthStatus(registry))

	// megahorn_auth
	s.AddTool(mcplib.Tool{
		Name:        "megahorn_auth",
		Description: "Initiate authentication for a platform (opens browser for user to log in)",
		InputSchema: mcplib.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"platform": map[string]any{"type": "string", "enum": []string{"twitter", "linkedin", "reddit"}},
				"headed":   map[string]any{"type": "boolean", "default": true},
			},
			Required: []string{"platform"},
		},
	}, handleAuth(registry))
}

func handlePost(registry *platform.Registry) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		args := req.GetArguments()
		platformName, _ := args["platform"].(string)
		content, _ := args["content"].(string)
		subreddit, _ := args["subreddit"].(string)
		dryRun, _ := args["dry_run"].(bool)

		p, ok := registry.Get(platformName)
		if !ok {
			return mcplib.NewToolResultError(fmt.Sprintf("unknown platform: %s", platformName)), nil
		}

		var subreddits []string
		if subreddit != "" {
			for _, s := range splitComma(subreddit) {
				subreddits = append(subreddits, s)
			}
		}

		result, err := p.Post(content, platform.PostOpts{
			Subreddits: subreddits,
			DryRun:     dryRun,
		})
		if err != nil {
			return mcplib.NewToolResultError(err.Error()), nil
		}

		data, _ := json.Marshal(result)
		return mcplib.NewToolResultText(string(data)), nil
	}
}

func handleAuthStatus(registry *platform.Registry) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		statuses := make(map[string]string)
		for _, p := range registry.All() {
			statuses[p.Name()] = string(p.Status())
		}
		data, _ := json.Marshal(statuses)
		return mcplib.NewToolResultText(string(data)), nil
	}
}

func handleAuth(registry *platform.Registry) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		args := req.GetArguments()
		platformName, _ := args["platform"].(string)
		headed, ok := args["headed"].(bool)
		if !ok {
			headed = true
		}

		p, exists := registry.Get(platformName)
		if !exists {
			return mcplib.NewToolResultError(fmt.Sprintf("unknown platform: %s", platformName)), nil
		}

		err := p.Auth(platform.AuthOpts{Headed: headed})
		if err != nil {
			return mcplib.NewToolResultError(err.Error()), nil
		}

		return mcplib.NewToolResultText(fmt.Sprintf("Successfully authenticated with %s", platformName)), nil
	}
}

func splitComma(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
