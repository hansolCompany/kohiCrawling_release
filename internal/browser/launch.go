package browser

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func LaunchVisibleBrowser(pw *playwright.Playwright) (playwright.Browser, playwright.Page, error) {
	browserInstance, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args:     []string{"--start-maximized"},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("브라우저 실행 실패: %w", err)
	}

	context, err := browserInstance.NewContext(playwright.BrowserNewContextOptions{
		NoViewport: playwright.Bool(true),
	})
	if err != nil {
		browserInstance.Close()
		return nil, nil, fmt.Errorf("브라우저 컨텍스트 생성 실패: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		context.Close()
		browserInstance.Close()
		return nil, nil, fmt.Errorf("페이지 생성 실패: %w", err)
	}

	return browserInstance, page, nil
}
