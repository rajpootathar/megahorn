package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

type ChromeOpts struct {
	Headed     bool
	ChromePath string
}

func NewContext(ctx context.Context, opts ChromeOpts) (context.Context, context.CancelFunc) {
	allocOpts := chromedp.DefaultExecAllocatorOptions[:]

	if opts.Headed {
		allocOpts = append(allocOpts,
			chromedp.Flag("headless", false),
		)
	}

	if opts.ChromePath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(opts.ChromePath))
	}

	allocOpts = append(allocOpts,
		chromedp.WindowSize(1280, 800),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	taskCtx, taskCancel := chromedp.NewContext(allocCtx)

	return taskCtx, func() {
		taskCancel()
		allocCancel()
	}
}

// CaptureScreenshot saves a screenshot to /tmp/megahorn/ and returns the path.
func CaptureScreenshot(ctx context.Context) string {
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		return ""
	}
	dir := filepath.Join(os.TempDir(), "megahorn")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, fmt.Sprintf("error_%d.png", os.Getpid()))
	os.WriteFile(path, buf, 0644)
	return path
}
