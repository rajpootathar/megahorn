package cmd

import "testing"

func TestParseSubreddits(t *testing.T) {
	result := parseSubreddits("SaaS,startups,webdev")
	if len(result) != 3 {
		t.Errorf("expected 3 subreddits, got %d", len(result))
	}
	if result[0] != "SaaS" {
		t.Errorf("expected SaaS, got %s", result[0])
	}
}

func TestParseSubredditsEmpty(t *testing.T) {
	result := parseSubreddits("")
	if len(result) != 0 {
		t.Errorf("expected 0 subreddits, got %d", len(result))
	}
}

func TestParseSubredditsTrimSpaces(t *testing.T) {
	result := parseSubreddits("SaaS, startups , webdev")
	if result[1] != "startups" {
		t.Errorf("expected trimmed 'startups', got '%s'", result[1])
	}
}
