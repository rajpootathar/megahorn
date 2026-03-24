package platform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLinkedInName(t *testing.T) {
	l := NewLinkedIn(nil, nil)
	if l.Name() != "linkedin" {
		t.Errorf("expected linkedin, got %s", l.Name())
	}
}

func TestLinkedInStatusNotConfigured(t *testing.T) {
	l := NewLinkedIn(nil, nil)
	if l.Status() != AuthStatusNotConfigured {
		t.Errorf("expected not_configured")
	}
}

func TestLinkedInPostDryRun(t *testing.T) {
	l := NewLinkedIn(nil, nil)
	result, err := l.Post("test post", PostOpts{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry run should succeed")
	}
}

func TestLinkedInPostMocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/posts":
			w.Header().Set("x-restli-id", "urn:li:share:12345")
			w.WriteHeader(http.StatusCreated)
		case "/v2/userinfo":
			json.NewEncoder(w).Encode(map[string]any{"sub": "abc123"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	l := NewLinkedIn(nil, nil)
	l.baseURL = srv.URL
	l.httpClient = srv.Client()

	result, err := l.postWithToken("test post", PostOpts{}, "fake-token", "urn:li:person:abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}
