package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Browser   BrowserConfig   `mapstructure:"browser"`
	Platforms PlatformsConfig `mapstructure:"platforms"`
	Reddit    RedditConfig    `mapstructure:"reddit"`
}

type BrowserConfig struct {
	Headed     bool   `mapstructure:"headed"`
	ChromePath string `mapstructure:"chrome_path"`
}

type PlatformsConfig struct {
	Defaults []string `mapstructure:"defaults"`
}

type RedditConfig struct {
	DefaultSubreddits []string `mapstructure:"default_subreddits"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".megahorn")
}

func DefaultPath() string {
	return filepath.Join(Dir(), "config.yaml")
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	v.SetDefault("browser.headed", false)
	v.SetDefault("browser.chrome_path", "")
	v.SetDefault("platforms.defaults", []string{})
	v.SetDefault("reddit.default_subreddits", []string{})

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				if !os.IsNotExist(err) {
					return nil, err
				}
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SetValue(path, key, value string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	v.ReadInConfig()

	if value == "true" || value == "false" {
		b, _ := strconv.ParseBool(value)
		v.Set(key, b)
	} else if strings.Contains(value, ",") {
		v.Set(key, strings.Split(value, ","))
	} else {
		v.Set(key, value)
	}

	return v.WriteConfig()
}

func EnsureDir() error {
	return os.MkdirAll(Dir(), 0700)
}
