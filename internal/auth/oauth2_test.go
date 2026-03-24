package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCallbackServerStartsAndStops(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	srv := NewCallbackServer(0)
	go srv.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	port := srv.Port()
	if port == 0 {
		t.Fatal("expected non-zero port")
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/callback?code=testcode&state=teststate", port))
	if err != nil {
		t.Fatalf("failed to hit callback: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	result := <-srv.Result()
	if result.Code != "testcode" {
		t.Errorf("expected testcode, got %s", result.Code)
	}
}

func TestCallbackServerRedirectURI(t *testing.T) {
	srv := NewCallbackServer(8338)
	if srv.RedirectURI() != "http://localhost:8338/callback" {
		t.Errorf("unexpected redirect URI: %s", srv.RedirectURI())
	}
}
