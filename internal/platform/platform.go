package platform

type Platform interface {
	Name() string
	Auth(opts AuthOpts) error
	Post(content string, opts PostOpts) (*PostResult, error)
	Status() AuthStatus
}

type AuthOpts struct {
	Headed bool
}

type PostOpts struct {
	Subreddits []string
	DryRun     bool
	Headed     bool
}

type PostResult struct {
	URL      string `json:"url"`
	Platform string `json:"platform"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

type AuthStatus string

const (
	AuthStatusAuthenticated AuthStatus = "authenticated"
	AuthStatusExpired       AuthStatus = "expired"
	AuthStatusNotConfigured AuthStatus = "not_configured"
)
