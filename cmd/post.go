package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rajpootathar/megahorn/internal/auth"
	"github.com/rajpootathar/megahorn/internal/config"
	"github.com/rajpootathar/megahorn/internal/platform"
	"github.com/spf13/cobra"
)

var (
	postTwitter   bool
	postLinkedIn  bool
	postReddit    bool
	postAll       bool
	postSubreddit string
	postFile      string
	postDryRun    bool
	postJSON      bool
	postHeaded    bool
)

func parseSubreddits(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

var postCmd = &cobra.Command{
	Use:   "post [content]",
	Short: "Post content to social media platforms",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var content string
		if postFile != "" {
			data, err := os.ReadFile(postFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			content = string(data)
		} else if len(args) > 0 {
			content = args[0]
		} else {
			return fmt.Errorf("provide content as argument or use --file")
		}

		cfg, _ := config.Load(config.DefaultPath())
		kr := auth.NewKeyring()

		registry := platform.NewRegistry()
		registry.Register(platform.NewTwitter(kr, cfg))
		registry.Register(platform.NewLinkedIn(kr, cfg))
		registry.Register(platform.NewReddit(kr, cfg))

		var targets []string
		if postAll {
			for _, p := range registry.Authenticated() {
				targets = append(targets, p.Name())
			}
		} else if postTwitter || postLinkedIn || postReddit {
			if postTwitter {
				targets = append(targets, "twitter")
			}
			if postLinkedIn {
				targets = append(targets, "linkedin")
			}
			if postReddit {
				targets = append(targets, "reddit")
			}
		} else {
			targets = cfg.Platforms.Defaults
		}

		if len(targets) == 0 {
			return fmt.Errorf("no platforms selected. Use -t, -l, -r, --all, or set defaults in config")
		}

		opts := platform.PostOpts{
			Subreddits: parseSubreddits(postSubreddit),
			DryRun:     postDryRun,
			Headed:     postHeaded,
		}

		var results []*platform.PostResult
		for _, name := range targets {
			p, ok := registry.Get(name)
			if !ok {
				fmt.Fprintf(os.Stderr, "Unknown platform: %s\n", name)
				continue
			}

			result, err := p.Post(content, opts)
			if err != nil {
				r := &platform.PostResult{Platform: name, Success: false, Error: err.Error()}
				results = append(results, r)
				if !postJSON {
					fmt.Fprintf(os.Stderr, "%s: ERROR — %v\n", name, err)
				}
				continue
			}

			results = append(results, result)
			if !postJSON {
				if result.Success {
					fmt.Printf("%s: %s\n", strings.ToUpper(name), result.URL)
				} else {
					fmt.Fprintf(os.Stderr, "%s: FAILED — %s\n", name, result.Error)
				}
			}
		}

		if postJSON {
			out, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		var succeeded, failed int
		for _, r := range results {
			if r.Success {
				succeeded++
			} else {
				failed++
			}
		}
		fmt.Printf("\n%d/%d published.", succeeded, len(results))
		if failed > 0 {
			fmt.Printf(" %d failed.", failed)
		}
		fmt.Println()

		return nil
	},
}

func init() {
	postCmd.Flags().BoolVarP(&postTwitter, "twitter", "t", false, "Post to Twitter")
	postCmd.Flags().BoolVarP(&postLinkedIn, "linkedin", "l", false, "Post to LinkedIn")
	postCmd.Flags().BoolVarP(&postReddit, "reddit", "r", false, "Post to Reddit")
	postCmd.Flags().BoolVarP(&postAll, "all", "a", false, "Post to all authenticated platforms")
	postCmd.Flags().StringVar(&postSubreddit, "subreddit", "", "Target subreddit(s), comma-separated")
	postCmd.Flags().StringVarP(&postFile, "file", "f", "", "Read content from file")
	postCmd.Flags().BoolVar(&postDryRun, "dry-run", false, "Preview without posting")
	postCmd.Flags().BoolVar(&postJSON, "json", false, "Output as JSON")
	postCmd.Flags().BoolVar(&postHeaded, "headed", false, "Use visible browser for Twitter")
	rootCmd.AddCommand(postCmd)
}
