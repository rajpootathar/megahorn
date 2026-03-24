package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type CallbackResult struct {
	Code  string
	State string
	Error string
}

type CallbackServer struct {
	port       int
	actualPort int
	result     chan CallbackResult
	listener   net.Listener
}

func NewCallbackServer(port int) *CallbackServer {
	return &CallbackServer{
		port:   port,
		result: make(chan CallbackResult, 1),
	}
}

func (s *CallbackServer) Port() int {
	if s.actualPort != 0 {
		return s.actualPort
	}
	return s.port
}

func (s *CallbackServer) Result() <-chan CallbackResult {
	return s.result
}

func (s *CallbackServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errMsg := r.URL.Query().Get("error")

		s.result <- CallbackResult{Code: code, State: state, Error: errMsg}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Megahorn</h1><p>Authorization complete. You can close this tab.</p></body></html>")
	})

	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	s.actualPort = s.listener.Addr().(*net.TCPAddr).Port

	srv := &http.Server{Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	return srv.Serve(s.listener)
}

func (s *CallbackServer) RedirectURI() string {
	return fmt.Sprintf("http://localhost:%d/callback", s.Port())
}

// RunOAuth2Flow performs the full OAuth2 authorization code flow:
// 1. Starts local callback server
// 2. Opens browser to authorization URL
// 3. Waits for callback with auth code
// 4. Exchanges code for token
func RunOAuth2Flow(cfg *oauth2.Config) (*oauth2.Token, error) {
	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	srv := NewCallbackServer(8338)
	cfg.RedirectURL = srv.RedirectURI()

	go srv.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Opening browser for authorization...\n")
	if err := OpenBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser. Please visit:\n%s\n", authURL)
	}

	fmt.Println("Waiting for authorization (timeout: 2 minutes)...")
	select {
	case result := <-srv.Result():
		if result.Error != "" {
			return nil, fmt.Errorf("authorization error: %s", result.Error)
		}
		if result.State != state {
			return nil, fmt.Errorf("state mismatch — possible CSRF")
		}
		token, err := cfg.Exchange(ctx, result.Code)
		if err != nil {
			return nil, fmt.Errorf("token exchange failed: %w", err)
		}
		return token, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("authorization timed out")
	}
}
