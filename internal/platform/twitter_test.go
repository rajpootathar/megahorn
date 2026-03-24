package platform

import "testing"

func TestTwitterName(t *testing.T) {
	tw := NewTwitter(nil, nil)
	if tw.Name() != "twitter" {
		t.Errorf("expected twitter, got %s", tw.Name())
	}
}

func TestTwitterStatusNotConfigured(t *testing.T) {
	tw := NewTwitter(nil, nil)
	if tw.Status() != AuthStatusNotConfigured {
		t.Errorf("expected not_configured")
	}
}

func TestTwitterPostDryRun(t *testing.T) {
	tw := NewTwitter(nil, nil)
	result, err := tw.Post("test tweet", PostOpts{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("dry run should succeed")
	}
	if result.Platform != "twitter" {
		t.Errorf("expected twitter, got %s", result.Platform)
	}
}

func TestTwitterPostRequiresAuth(t *testing.T) {
	tw := NewTwitter(nil, nil)
	_, err := tw.Post("test", PostOpts{})
	if err == nil {
		t.Error("expected error when not authenticated")
	}
}
