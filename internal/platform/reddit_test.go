package platform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedditName(t *testing.T) {
	r := NewReddit(nil, nil)
	if r.Name() != "reddit" {
		t.Errorf("expected reddit, got %s", r.Name())
	}
}

func TestRedditStatusNotConfigured(t *testing.T) {
	r := NewReddit(nil, nil)
	if r.Status() != AuthStatusNotConfigured {
		t.Errorf("expected not_configured, got %s", r.Status())
	}
}

func TestRedditPostDryRun(t *testing.T) {
	r := NewReddit(nil, nil)
	result, err := r.Post("test post", PostOpts{DryRun: true, Subreddits: []string{"test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry run should succeed")
	}
	if result.Platform != "reddit" {
		t.Errorf("expected reddit, got %s", result.Platform)
	}
}

func TestRedditPostDryRunMultiSubreddit(t *testing.T) {
	r := NewReddit(nil, nil)
	result, err := r.Post("test post", PostOpts{DryRun: true, Subreddits: []string{"SaaS", "startups"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry run should succeed")
	}
}

func TestRedditPostRequiresSubreddit(t *testing.T) {
	r := NewReddit(nil, nil)
	_, err := r.Post("test", PostOpts{})
	if err == nil {
		t.Error("expected error when no subreddit provided")
	}
}

func TestRedditPostMocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/submit":
			json.NewEncoder(w).Encode(map[string]any{
				"json": map[string]any{
					"data": map[string]any{
						"url": "https://reddit.com/r/test/comments/abc123",
					},
				},
			})
		case "/r/test/about.json":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"display_name": "test"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	r := NewReddit(nil, nil)
	r.baseURL = srv.URL
	r.httpClient = srv.Client()

	result, err := r.postToSubreddit("test post", "test", "fake-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.URL != "https://reddit.com/r/test/comments/abc123" {
		t.Errorf("unexpected URL: %s", result.URL)
	}
}

func TestRedditSubredditValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/r/nonexistent/about.json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer srv.Close()

	r := NewReddit(nil, nil)
	r.baseURL = srv.URL
	r.httpClient = srv.Client()

	err := r.validateSubreddit("nonexistent", "fake-token")
	if err == nil {
		t.Error("expected error for nonexistent subreddit")
	}
}
