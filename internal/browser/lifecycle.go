package browser

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// EnsurePlaywrightBrowser는 Playwright 드라이버와 Chromium을 설치한다.
func EnsurePlaywrightBrowser() error {
	fmt.Println("Playwright Chromium 설치 확인 중...")

	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		return fmt.Errorf("Playwright Chromium 설치 실패: %w", err)
	}

	fmt.Println("Playwright Chromium 준비 완료")
	return nil
}

// IsClosed reports whether the browser page or its browser is no longer available.
func IsClosed(page playwright.Page) bool {
	return isBrowserClosed(page)
}

// WaitForBrowserClose는 사용자가 브라우저를 직접 닫을 때까지 대기한다.
func WaitForBrowserClose(page playwright.Page) {
	for {
		time.Sleep(1 * time.Second)

		if isBrowserClosed(page) {
			fmt.Println("브라우저가 종료되었습니다.")
			break
		}
	}
}

func isBrowserClosed(page playwright.Page) bool {
	defer func() { recover() }()

	if page == nil || page.IsClosed() {
		return true
	}

	ctx := page.Context()
	if ctx == nil {
		return true
	}

	browser := ctx.Browser()
	if browser == nil || !browser.IsConnected() {
		return true
	}

	return false
}
