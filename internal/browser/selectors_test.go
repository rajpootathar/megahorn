package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTwitterSelectors(t *testing.T) {
	s := DefaultTwitterSelectors()
	if s.ComposeButton == "" {
		t.Error("expected compose button selector")
	}
	if s.TweetTextarea == "" {
		t.Error("expected tweet textarea selector")
	}
	if s.PostButton == "" {
		t.Error("expected post button selector")
	}
}

func TestLoadSelectorsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "twitter.yaml")
	content := []byte("compose_button: \"[data-testid=custom]\"\ntweet_textarea: \"[data-testid=custom2]\"\npost_button: \"[data-testid=custom3]\"\n")
	os.WriteFile(path, content, 0644)

	s, err := LoadTwitterSelectors(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ComposeButton != "[data-testid=custom]" {
		t.Errorf("unexpected selector: %s", s.ComposeButton)
	}
}

func TestResolveSelectorsWithOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "twitter.yaml")
	content := []byte("tweet_textarea: \"[data-testid=override]\"\n")
	os.WriteFile(path, content, 0644)

	s := ResolveTwitterSelectors(path)
	if s.TweetTextarea != "[data-testid=override]" {
		t.Errorf("expected override, got %s", s.TweetTextarea)
	}
	if s.ComposeButton == "" {
		t.Error("compose button should fall back to default")
	}
}
