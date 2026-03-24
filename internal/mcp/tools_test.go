package mcp

import "testing"

func TestToolCount(t *testing.T) {
	tools := BuildToolList()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}
}

func TestToolNames(t *testing.T) {
	tools := BuildToolList()
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	expected := []string{"megahorn_post", "megahorn_auth_status", "megahorn_auth"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}
