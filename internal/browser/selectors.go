package browser

import (
	"os"
	"path/filepath"

	"github.com/rajpootathar/megahorn/internal/config"
	"gopkg.in/yaml.v3"
)

type TwitterSelectors struct {
	ComposeButton string `yaml:"compose_button"`
	TweetTextarea string `yaml:"tweet_textarea"`
	PostButton    string `yaml:"post_button"`
	TweetLink     string `yaml:"tweet_link"`
}

func DefaultTwitterSelectors() TwitterSelectors {
	return TwitterSelectors{
		ComposeButton: `[data-testid="SideNav_NewTweet_Button"]`,
		TweetTextarea: `[data-testid="tweetTextarea_0"]`,
		PostButton:    `[data-testid="tweetButtonInline"]`,
		TweetLink:     `a[href*="/status/"]`,
	}
}

func LoadTwitterSelectors(path string) (*TwitterSelectors, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s TwitterSelectors
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func ResolveTwitterSelectors(overridePath string) TwitterSelectors {
	defaults := DefaultTwitterSelectors()

	override, err := LoadTwitterSelectors(overridePath)
	if err != nil {
		return defaults
	}

	if override.ComposeButton != "" {
		defaults.ComposeButton = override.ComposeButton
	}
	if override.TweetTextarea != "" {
		defaults.TweetTextarea = override.TweetTextarea
	}
	if override.PostButton != "" {
		defaults.PostButton = override.PostButton
	}
	if override.TweetLink != "" {
		defaults.TweetLink = override.TweetLink
	}
	return defaults
}

func UserSelectorsPath() string {
	return filepath.Join(config.Dir(), "selectors", "twitter.yaml")
}
