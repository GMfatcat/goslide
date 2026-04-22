package pdfexport

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ChromedpLauncher drives a real Chrome instance via chromedp.
type ChromedpLauncher struct{}

// NewChromedpLauncher returns a Launcher that uses chromedp.
func NewChromedpLauncher() *ChromedpLauncher {
	return &ChromedpLauncher{}
}

func (l *ChromedpLauncher) Launch(ctx context.Context, req LaunchRequest) ([]byte, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(req.ChromePath),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	// 60-second overall timeout; Chrome launch + page load is typically < 5s.
	tCtx, cancelT := context.WithTimeout(browserCtx, 60*time.Second)
	defer cancelT()

	var pdfData []byte
	err := chromedp.Run(tCtx,
		chromedp.Navigate(req.URL),
		chromedp.WaitReady(".reveal", chromedp.ByQuery),
		chromedp.Poll("window.__goslideReady === true", nil,
			chromedp.WithPollingInterval(100*time.Millisecond),
			chromedp.WithPollingTimeout(30*time.Second),
		),
		chromedp.ActionFunc(func(ctx context.Context) error {
			params := page.PrintToPDFParams{
				PaperWidth:        req.PaperWidthIn,
				PaperHeight:       req.PaperHeightIn,
				PrintBackground:   true,
				PreferCSSPageSize: true,
				MarginTop:         0,
				MarginBottom:      0,
				MarginLeft:        0,
				MarginRight:       0,
			}
			b, _, err := params.Do(ctx)
			if err != nil {
				return err
			}
			pdfData = b
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return pdfData, nil
}
