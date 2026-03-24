package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	authpkg "github.com/rajpootathar/megahorn/internal/auth"
	"github.com/rajpootathar/megahorn/internal/config"
	"golang.org/x/oauth2"
)

type LinkedIn struct {
	keyring    *authpkg.Keyring
	config     *config.Config
	baseURL    string
	httpClient *http.Client
}

func NewLinkedIn(kr *authpkg.Keyring, cfg *config.Config) *LinkedIn {
	return &LinkedIn{
		keyring:    kr,
		config:     cfg,
		baseURL:    "https://api.linkedin.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (l *LinkedIn) Name() string { return "linkedin" }

func (l *LinkedIn) Status() AuthStatus {
	if l.keyring == nil {
		return AuthStatusNotConfigured
	}
	_, err := l.keyring.Get("linkedin", "access_token")
	if err != nil {
		_, err2 := l.keyring.Get("linkedin", "client_id")
		if err2 != nil {
			return AuthStatusNotConfigured
		}
		return AuthStatusExpired
	}
	expiryStr, err := l.keyring.Get("linkedin", "token_expiry")
	if err == nil {
		expiry, err := time.Parse(time.RFC3339, expiryStr)
		if err == nil && time.Now().After(expiry) {
			return AuthStatusExpired
		}
		if err == nil && time.Until(expiry) < 7*24*time.Hour {
			fmt.Printf("WARNING: LinkedIn token expires in %d days. Re-run: megahorn auth linkedin\n", int(time.Until(expiry).Hours()/24))
		}
	}
	return AuthStatusAuthenticated
}

func (l *LinkedIn) Auth(opts AuthOpts) error {
	fmt.Println("LinkedIn OAuth2 Setup")
	fmt.Println("1. Go to https://www.linkedin.com/developers/apps/new")
	fmt.Println("2. Create an app (any name, e.g., 'Megahorn')")
	fmt.Println("3. Under Products, request 'Share on LinkedIn' and 'Sign In with LinkedIn using OpenID Connect'")
	fmt.Println("4. Under Auth settings, add redirect URL: http://localhost:8338/callback")
	fmt.Println("5. Copy your Client ID and Client Secret")
	fmt.Println()

	var clientID, clientSecret string
	fmt.Print("Client ID: ")
	fmt.Scanln(&clientID)
	fmt.Print("Client Secret: ")
	fmt.Scanln(&clientSecret)

	if l.keyring != nil {
		l.keyring.Set("linkedin", "client_id", clientID)
		l.keyring.Set("linkedin", "client_secret", clientSecret)
	}

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile", "w_member_social"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
		},
	}

	token, err := authpkg.RunOAuth2Flow(oauthCfg)
	if err != nil {
		return fmt.Errorf("linkedin auth failed: %w", err)
	}

	if l.keyring != nil {
		l.keyring.Set("linkedin", "access_token", token.AccessToken)
		l.keyring.Set("linkedin", "token_expiry", token.Expiry.Format(time.RFC3339))
	}

	personURN, err := l.fetchPersonURN(token.AccessToken)
	if err != nil {
		fmt.Printf("Warning: could not fetch profile URN: %v\n", err)
	} else if l.keyring != nil {
		l.keyring.Set("linkedin", "person_urn", personURN)
	}

	fmt.Println("LinkedIn authentication successful!")
	return nil
}

func (l *LinkedIn) fetchPersonURN(token string) (string, error) {
	req, _ := http.NewRequest("GET", l.baseURL+"/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Sub string `json:"sub"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Sub == "" {
		return "", fmt.Errorf("empty sub in userinfo response")
	}
	return fmt.Sprintf("urn:li:person:%s", result.Sub), nil
}

func (l *LinkedIn) Post(content string, opts PostOpts) (*PostResult, error) {
	if opts.DryRun {
		return &PostResult{
			Platform: "linkedin",
			Success:  true,
			URL:      "[DRY RUN] would post to LinkedIn",
		}, nil
	}

	if l.keyring == nil {
		return nil, fmt.Errorf("not authenticated — run: megahorn auth linkedin")
	}

	token, err := l.keyring.Get("linkedin", "access_token")
	if err != nil {
		return nil, fmt.Errorf("not authenticated — run: megahorn auth linkedin")
	}

	expiryStr, _ := l.keyring.Get("linkedin", "token_expiry")
	if expiryStr != "" {
		expiry, _ := time.Parse(time.RFC3339, expiryStr)
		if time.Now().After(expiry) {
			return nil, fmt.Errorf("LinkedIn token expired — run: megahorn auth linkedin")
		}
	}

	personURN, err := l.keyring.Get("linkedin", "person_urn")
	if err != nil {
		return nil, fmt.Errorf("missing person URN — re-run: megahorn auth linkedin")
	}

	return l.postWithToken(content, opts, token, personURN)
}

func (l *LinkedIn) postWithToken(content string, opts PostOpts, token, personURN string) (*PostResult, error) {
	payload := map[string]any{
		"author":         personURN,
		"lifecycleState": "PUBLISHED",
		"visibility":     "PUBLIC",
		"distribution": map[string]any{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []any{},
			"thirdPartyDistributionChannels": []any{},
		},
		"commentary": content,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", l.baseURL+"/rest/posts", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("LinkedIn-Version", "202401")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linkedin API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return &PostResult{
			Platform: "linkedin",
			Success:  false,
			Error:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	shareID := resp.Header.Get("x-restli-id")
	postURL := fmt.Sprintf("https://www.linkedin.com/feed/update/%s", shareID)

	return &PostResult{
		Platform: "linkedin",
		Success:  true,
		URL:      postURL,
	}, nil
}
