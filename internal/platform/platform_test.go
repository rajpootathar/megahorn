package platform

import "testing"

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	platforms := r.All()
	if len(platforms) != 0 {
		t.Errorf("expected empty registry, got %d", len(platforms))
	}
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockPlatform{name: "twitter"})

	p, ok := r.Get("twitter")
	if !ok {
		t.Fatal("expected to find twitter")
	}
	if p.Name() != "twitter" {
		t.Errorf("expected twitter, got %s", p.Name())
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestRegistryAuthenticated(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockPlatform{name: "a", status: AuthStatusAuthenticated})
	r.Register(&mockPlatform{name: "b", status: AuthStatusNotConfigured})
	auth := r.Authenticated()
	if len(auth) != 1 {
		t.Errorf("expected 1 authenticated, got %d", len(auth))
	}
}

type mockPlatform struct {
	name   string
	status AuthStatus
}

func (m *mockPlatform) Name() string                                          { return m.name }
func (m *mockPlatform) Auth(opts AuthOpts) error                              { return nil }
func (m *mockPlatform) Post(content string, opts PostOpts) (*PostResult, error) {
	return &PostResult{URL: "https://example.com", Platform: m.name, Success: true}, nil
}
func (m *mockPlatform) Status() AuthStatus {
	if m.status == "" {
		return AuthStatusNotConfigured
	}
	return m.status
}
