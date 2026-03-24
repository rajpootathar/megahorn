package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/network"
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
	_, err := tw.keyring.Get("twitter", "cookies")
	if err != nil {
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
	return browser.ChromeOpts{
		Headed:     headed,
		ChromePath: cfg.Browser.ChromePath,
	}
}

func (tw *Twitter) Auth(opts AuthOpts) error {
	headed := opts.Headed
	ctx, cancel := browser.NewContext(context.Background(), tw.chromeOpts(&headed))
	defer cancel()

	fmt.Println("Opening Twitter login page...")
	fmt.Println("Please log in manually (handle 2FA if needed).")
	fmt.Println("Press Enter here once you're logged in and see your feed...")

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://x.com/login"),
	)
	if err != nil {
		return fmt.Errorf("failed to open Twitter: %w", err)
	}

	fmt.Scanln()

	// Capture cookies as JSON
	var cookies []*network.Cookie
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to capture cookies: %w", err)
	}

	cookieJSON, err := json.Marshal(cookies)
	if err != nil {
		return fmt.Errorf("failed to serialize cookies: %w", err)
	}

	if tw.keyring != nil {
		tw.keyring.Set("twitter", "cookies", string(cookieJSON))
	}

	fmt.Println("Twitter auth saved.")
	return nil
}

func (tw *Twitter) restoreCookies(ctx context.Context) error {
	if tw.keyring == nil {
		return fmt.Errorf("no keyring")
	}
	cookieJSON, err := tw.keyring.Get("twitter", "cookies")
	if err != nil {
		return fmt.Errorf("no saved cookies — run: megahorn auth twitter")
	}

	var cookies []*network.Cookie
	if err := json.Unmarshal([]byte(cookieJSON), &cookies); err != nil {
		return fmt.Errorf("corrupt cookies — re-run: megahorn auth twitter")
	}

	return chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, c := range cookies {
				err := network.SetCookie(c.Name, c.Value).
					WithDomain(c.Domain).
					WithPath(c.Path).
					WithHTTPOnly(c.HTTPOnly).
					WithSecure(c.Secure).
					Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
	)
}

func (tw *Twitter) Post(content string, opts PostOpts) (*PostResult, error) {
	if opts.DryRun {
		return &PostResult{
			Platform: "twitter",
			Success:  true,
			URL:      "[DRY RUN] would post to Twitter",
		}, nil
	}

	headed := opts.Headed
	ctx, cancel := browser.NewContext(context.Background(), tw.chromeOpts(&headed))
	defer cancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, 90*time.Second)
	defer timeoutCancel()

	if err := tw.restoreCookies(ctx); err != nil {
		return nil, err
	}

	err := chromedp.Run(ctx,
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
