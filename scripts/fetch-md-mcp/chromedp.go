// Phase 2 (2026-05-09) — JS-rendering path via chromedp. Used when a
// tool's `js: true` param is set. Launches headless Chromium (system
// `chromium` binary on alpine; `chrome`/`chromium` on host), renders
// the page, returns the rendered DOM HTML for downstream readability +
// markdown conversion.
//
// Resource shape: each call spawns its own browser allocator + tab.
// Higher latency than persistent-browser, but isolation is cleaner
// and simpler to reason about. Switch to a shared allocator if a
// multi-fetch batch becomes the hot path.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// fetchURLJS renders `target` in headless Chromium and returns the
// rendered HTML. waitMS is the post-load settle (default 500ms) to
// catch SPA hydration. Honors ctx deadline.
func fetchURLJS(ctx context.Context, cfg *fetchConfig, target string, waitMS int) (string, string, error) {
	if waitMS <= 0 {
		waitMS = 500
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.UserAgent(cfg.UserAgent),
		)...)
	defer allocCancel()

	tabCtx, tabCancel := chromedp.NewContext(allocCtx)
	defer tabCancel()

	var (
		html     string
		finalURL string
	)
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(target),
		chromedp.Sleep(time.Duration(waitMS)*time.Millisecond),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.Location(&finalURL),
	)
	if err != nil {
		return "", target, fmt.Errorf("chromedp.Run: %w", err)
	}
	return html, finalURL, nil
}
