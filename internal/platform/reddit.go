package platform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rajpootathar/megahorn/internal/auth"
	"github.com/rajpootathar/megahorn/internal/config"
	"golang.org/x/oauth2"
)

type Reddit struct {
	keyring    *auth.Keyring
	config     *config.Config
	baseURL    string
	httpClient *http.Client
}

func NewReddit(kr *auth.Keyring, cfg *config.Config) *Reddit {
	return &Reddit{
		keyring:    kr,
		config:     cfg,
		baseURL:    "https://oauth.reddit.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *Reddit) Name() string { return "reddit" }

func (r *Reddit) Status() AuthStatus {
	if r.keyring == nil {
		return AuthStatusNotConfigured
	}
	_, err := r.keyring.Get("reddit", "access_token")
	if err != nil {
		_, err2 := r.keyring.Get("reddit", "client_id")
		if err2 != nil {
			return AuthStatusNotConfigured
		}
		return AuthStatusExpired
	}
	expiryStr, err := r.keyring.Get("reddit", "token_expiry")
	if err == nil {
		expiry, err := time.Parse(time.RFC3339, expiryStr)
		if err == nil && time.Now().After(expiry) {
			if _, refreshErr := r.refreshToken(); refreshErr != nil {
				return AuthStatusExpired
			}
		}
	}
	return AuthStatusAuthenticated
}

func (r *Reddit) Auth(opts AuthOpts) error {
	fmt.Println("Reddit OAuth2 Setup")
	fmt.Println("1. Go to https://www.reddit.com/prefs/apps")
	fmt.Println("2. Create a 'web app' (redirect URI: http://localhost:8338/callback)")
	fmt.Println("3. Copy your Client ID (under app name) and Secret")
	fmt.Println()

	var clientID, clientSecret string
	fmt.Print("Client ID: ")
	fmt.Scanln(&clientID)
	fmt.Print("Client Secret: ")
	fmt.Scanln(&clientSecret)

	if r.keyring != nil {
		r.keyring.Set("reddit", "client_id", clientID)
		r.keyring.Set("reddit", "client_secret", clientSecret)
	}

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"submit", "read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.reddit.com/api/v1/authorize",
			TokenURL: "https://www.reddit.com/api/v1/access_token",
		},
	}

	token, err := auth.RunOAuth2Flow(oauthCfg)
	if err != nil {
		return fmt.Errorf("reddit auth failed: %w", err)
	}

	if r.keyring != nil {
		r.keyring.Set("reddit", "access_token", token.AccessToken)
		if token.RefreshToken != "" {
			r.keyring.Set("reddit", "refresh_token", token.RefreshToken)
		}
		r.keyring.Set("reddit", "token_expiry", token.Expiry.Format(time.RFC3339))
	}

	fmt.Println("Reddit authentication successful!")
	return nil
}

func (r *Reddit) refreshToken() (string, error) {
	if r.keyring == nil {
		return "", fmt.Errorf("no keyring")
	}
	refreshToken, err := r.keyring.Get("reddit", "refresh_token")
	if err != nil {
		return "", fmt.Errorf("no refresh token — run: megahorn auth reddit")
	}
	clientID, _ := r.keyring.Get("reddit", "client_id")
	clientSecret, _ := r.keyring.Get("reddit", "client_secret")

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "megahorn/0.1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.AccessToken == "" {
		return "", fmt.Errorf("refresh failed")
	}

	expiry := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	r.keyring.Set("reddit", "access_token", result.AccessToken)
	r.keyring.Set("reddit", "token_expiry", expiry.Format(time.RFC3339))

	return result.AccessToken, nil
}

func (r *Reddit) getToken() (string, error) {
	if r.keyring == nil {
		return "", fmt.Errorf("not authenticated — run: megahorn auth reddit")
	}
	token, err := r.keyring.Get("reddit", "access_token")
	if err != nil {
		return "", fmt.Errorf("not authenticated — run: megahorn auth reddit")
	}
	expiryStr, _ := r.keyring.Get("reddit", "token_expiry")
	if expiryStr != "" {
		expiry, _ := time.Parse(time.RFC3339, expiryStr)
		if time.Now().After(expiry) {
			return r.refreshToken()
		}
	}
	return token, nil
}

func (r *Reddit) Post(content string, opts PostOpts) (*PostResult, error) {
	if len(opts.Subreddits) == 0 {
		return nil, fmt.Errorf("subreddit is required — use --subreddit")
	}

	if opts.DryRun {
		return &PostResult{
			Platform: "reddit",
			Success:  true,
			URL:      fmt.Sprintf("[DRY RUN] would post to r/%s", strings.Join(opts.Subreddits, ", r/")),
		}, nil
	}

	token, err := r.getToken()
	if err != nil {
		return nil, err
	}

	var lastResult *PostResult
	var errors []string
	for _, sub := range opts.Subreddits {
		result, err := r.postToSubreddit(content, sub, token)
		if err != nil {
			errors = append(errors, fmt.Sprintf("r/%s: %v", sub, err))
			continue
		}
		lastResult = result
		fmt.Printf("  r/%s: %s\n", sub, result.URL)
	}

	if len(errors) > 0 && lastResult == nil {
		return &PostResult{
			Platform: "reddit",
			Success:  false,
			Error:    strings.Join(errors, "; "),
		}, nil
	}

	if lastResult == nil {
		lastResult = &PostResult{Platform: "reddit", Success: true}
	}
	return lastResult, nil
}

func (r *Reddit) validateSubreddit(subreddit, token string) error {
	req, _ := http.NewRequest("GET", r.baseURL+"/r/"+subreddit+"/about.json", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "megahorn/0.1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subreddit r/%s not found or inaccessible", subreddit)
	}
	return nil
}

func (r *Reddit) postToSubreddit(content, subreddit, token string) (*PostResult, error) {
	if err := r.validateSubreddit(subreddit, token); err != nil {
		return nil, err
	}

	lines := strings.SplitN(content, "\n", 2)
	title := lines[0]
	body := ""
	if len(lines) > 1 {
		body = strings.TrimSpace(lines[1])
	}

	data := url.Values{
		"sr":    {subreddit},
		"kind":  {"self"},
		"title": {title},
		"text":  {body},
	}

	req, _ := http.NewRequest("POST", r.baseURL+"/api/submit", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "megahorn/0.1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit API error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &PostResult{
			Platform: "reddit",
			Success:  false,
			Error:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)),
		}, nil
	}

	var result struct {
		JSON struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"json"`
	}
	json.Unmarshal(bodyBytes, &result)

	return &PostResult{
		Platform: "reddit",
		Success:  true,
		URL:      result.JSON.Data.URL,
	}, nil
}
