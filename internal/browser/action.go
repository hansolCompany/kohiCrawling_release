package browser

import (
	"time"

	"github.com/playwright-community/playwright-go"
)

// WaitAfterAction은 클릭·페이지 이동 후 UI 안정화를 위해 1초 대기한다.
func WaitAfterAction() {
	time.Sleep(1 * time.Second)
}

func WaitForPageReady(page playwright.Page) (string, error) {
	msg, err := WaitLoading(page)
	if err != nil {
		return msg, err
	}
	WaitAfterAction()
	return msg, nil
}

func ClickVisible(locator playwright.Locator, timeoutMs float64) error {
	target := locator.First()
	if err := target.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(timeoutMs),
	}); err != nil {
		return err
	}
	if err := target.Click(); err != nil {
		return err
	}
	WaitAfterAction()
	return nil
}
