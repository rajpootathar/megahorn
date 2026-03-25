package platform

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	authpkg "github.com/rajpootathar/megahorn/internal/auth"
	"github.com/rajpootathar/megahorn/internal/browser"
	"github.com/rajpootathar/megahorn/internal/config"
)

type Twitter struct {
	keyring   *authpkg.Keyring
	config    *config.Config
	selectors browser.TwitterSelectors
}

func NewTwitter(kr *authpkg.Keyring, cfg *config.Config) *Twitter {
	selectors := browser.ResolveTwitterSelectors(browser.UserSelectorsPath())
	return &Twitter{
		keyring:   kr,
		config:    cfg,
		selectors: selectors,
	}
}

func (tw *Twitter) Name() string { return "twitter" }

func (tw *Twitter) Status() AuthStatus {
	if tw.keyring == nil {
		return AuthStatusNotConfigured
	}
	val, err := tw.keyring.Get("twitter", "auth")
	if err != nil || val != "true" {
		return AuthStatusNotConfigured
	}
	return AuthStatusAuthenticated
}

func (tw *Twitter) chromeOpts(headedOverride *bool) browser.ChromeOpts {
	cfg := &config.Config{}
	if tw.config != nil {
		cfg = tw.config
	}
	headed := cfg.Browser.Headed
	if headedOverride != nil {
		headed = *headedOverride
	}
	profileDir := cfg.Browser.ProfileDir
	if profileDir == "" {
		profileDir = config.DefaultProfileDir()
	}
	return browser.ChromeOpts{
		Headed:      headed,
		ChromePath:  cfg.Browser.ChromePath,
		UserDataDir: profileDir,
	}
}

func (tw *Twitter) Auth(opts AuthOpts) error {
	headed := opts.Headed
	ctx, cancel, err := browser.NewContext(context.Background(), tw.chromeOpts(&headed))
	if err != nil {
		return fmt.Errorf("failed to create browser context: %w", err)
	}
	defer cancel()

	fmt.Println("Opening Twitter login page...")
	fmt.Println("Please log in manually (handle 2FA if needed).")
	fmt.Println("Press Enter here once you're logged in and see your feed...")

	err = chromedp.Run(ctx,
		chromedp.Navigate("https://x.com/login"),
	)
	if err != nil {
		return fmt.Errorf("failed to open Twitter: %w", err)
	}

	fmt.Scanln()

	// Clean up old cookie-based auth from keyring (migration)
	if tw.keyring != nil {
		tw.keyring.Delete("twitter", "cookies")
		tw.keyring.Set("twitter", "auth", "true")
	}

	fmt.Println("Twitter auth saved.")
	return nil
}

func (tw *Twitter) Post(content string, opts PostOpts) (*PostResult, error) {
	if opts.DryRun {
		return &PostResult{
			Platform: "twitter",
			Success:  true,
			URL:      "[DRY RUN] would post to Twitter",
		}, nil
	}

	if tw.Status() != AuthStatusAuthenticated {
		return nil, fmt.Errorf("not authenticated — run: megahorn auth twitter")
	}

	headed := opts.Headed
	ctx, cancel, err := browser.NewContext(context.Background(), tw.chromeOpts(&headed))
	if err != nil {
		if strings.Contains(err.Error(), "lock") || strings.Contains(err.Error(), "already") {
			return nil, fmt.Errorf("another megahorn instance is running — close it and retry")
		}
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}
	defer cancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, 90*time.Second)
	defer timeoutCancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate("https://x.com/compose/tweet"),
		chromedp.WaitVisible(tw.selectors.TweetTextarea, chromedp.ByQuery),
	)
	if err != nil {
		screenshot := browser.CaptureScreenshot(ctx)
		errMsg := fmt.Sprintf("failed to open compose (selector: %s): %v", tw.selectors.TweetTextarea, err)
		if screenshot != "" {
			errMsg += fmt.Sprintf(" — screenshot saved to %s", screenshot)
		}
		return &PostResult{Platform: "twitter", Success: false, Error: errMsg}, nil
	}

	for _, ch := range content {
		err = chromedp.Run(ctx,
			chromedp.SendKeys(tw.selectors.TweetTextarea, string(ch), chromedp.ByQuery),
		)
		if err != nil {
			screenshot := browser.CaptureScreenshot(ctx)
			errMsg := fmt.Sprintf("failed to type: %v", err)
			if screenshot != "" {
				errMsg += fmt.Sprintf(" — screenshot: %s", screenshot)
			}
			return &PostResult{Platform: "twitter", Success: false, Error: errMsg}, nil
		}
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(tw.selectors.PostButton, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		screenshot := browser.CaptureScreenshot(ctx)
		errMsg := fmt.Sprintf("failed to post (selector: %s): %v", tw.selectors.PostButton, err)
		if screenshot != "" {
			errMsg += fmt.Sprintf(" — screenshot: %s", screenshot)
		}
		return &PostResult{Platform: "twitter", Success: false, Error: errMsg}, nil
	}

	return &PostResult{
		Platform: "twitter",
		Success:  true,
		URL:      "https://x.com (posted — URL extraction pending v1.1)",
	}, nil
}
