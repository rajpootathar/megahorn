package cmd

import (
	"fmt"
	"strings"

	"github.com/rajpootathar/megahorn/internal/auth"
	"github.com/rajpootathar/megahorn/internal/config"
	"github.com/rajpootathar/megahorn/internal/platform"
	"github.com/spf13/cobra"
)

var (
	authHeaded   bool
	authHeadless bool
)

func buildRegistry() *platform.Registry {
	cfg, _ := config.Load(config.DefaultPath())
	kr := auth.NewKeyring()

	registry := platform.NewRegistry()
	registry.Register(platform.NewTwitter(kr, cfg))
	registry.Register(platform.NewLinkedIn(kr, cfg))
	registry.Register(platform.NewReddit(kr, cfg))
	return registry
}

var authCmd = &cobra.Command{
	Use:   "auth [platform]",
	Short: "Authenticate with a social media platform",
	Long:  "Authenticate with twitter, linkedin, or reddit.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		registry := buildRegistry()

		p, ok := registry.Get(target)
		if !ok {
			return fmt.Errorf("unknown platform: %s (available: twitter, linkedin, reddit)", target)
		}

		headed := authHeaded
		if authHeadless {
			headed = false
		}

		return p.Auth(platform.AuthOpts{Headed: headed})
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status for all platforms",
	Run: func(cmd *cobra.Command, args []string) {
		registry := buildRegistry()

		fmt.Println("Platform       Status")
		fmt.Println(strings.Repeat("-", 30))
		for _, p := range registry.All() {
			status := p.Status()
			icon := "x"
			switch status {
			case platform.AuthStatusAuthenticated:
				icon = "+"
			case platform.AuthStatusExpired:
				icon = "!"
			}
			fmt.Printf("%-14s %s %s\n", p.Name(), icon, status)
		}
	},
}

func init() {
	authCmd.Flags().BoolVar(&authHeaded, "headed", true, "Use visible browser mode")
	authCmd.Flags().BoolVar(&authHeadless, "headless", false, "Use headless browser mode")
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
