package longterm

import (
	"fmt"
	"strings"
	"time"

	"kohiCrawling/internal/browser"

	"github.com/playwright-community/playwright-go"
)

const (
	nexacroLoadingSelector = `#mainframe\\.waitwindow\\:image`
	loadingCompleteTop     = "-49px"

	loadingMaxNotFound = 4
	loadingTimeout     = 90 * time.Second

	fastLoadingPollInterval = 50 * time.Millisecond
)

func WaitForLoading(page playwright.Page) error {
	return waitForLoading(page, 300*time.Millisecond, 500*time.Millisecond)
}

func WaitForLoadingFast(page playwright.Page) error {
	return waitForLoading(page, fastLoadingPollInterval, fastLoadingPollInterval)
}

func waitForLoading(page playwright.Page, initialWait, pollInterval time.Duration) error {
	browser.ClearDialogMessage()

	time.Sleep(initialWait)

	notFoundCount := 0
	deadline := time.Now().Add(loadingTimeout)

	for {
		if time.Now().After(deadline) {
			return nil
		}

		loading := page.Locator(nexacroLoadingSelector)
		count, err := loading.Count()
		if err != nil || count == 0 {
			notFoundCount++
			if notFoundCount >= loadingMaxNotFound {
				return nil
			}
			time.Sleep(pollInterval)
			continue
		}

		notFoundCount = 0

		style, err := loading.First().GetAttribute("style")
		if err != nil {
			notFoundCount++
			if notFoundCount >= loadingMaxNotFound {
				return nil
			}
			time.Sleep(pollInterval)
			continue
		}

		fmt.Printf("%s loadingStyle\n", style)
		if isNexacroLoadingComplete(style) {
			return nil
		}

		time.Sleep(pollInterval)
	}
}

func isNexacroLoadingComplete(style string) bool {
	parts := strings.Split(style, ";")
	if len(parts) == 0 {
		return false
	}

	first := strings.SplitN(parts[0], ":", 2)
	if len(first) != 2 {
		return false
	}

	return strings.TrimSpace(first[1]) == loadingCompleteTop
}

func ClickWithLoading(page playwright.Page, locator playwright.Locator, timeoutMs float64) error {
	return clickWithLoading(page, locator, timeoutMs, WaitForLoading)
}

func ClickWithLoadingFast(page playwright.Page, locator playwright.Locator, timeoutMs float64) error {
	return clickWithLoading(page, locator, timeoutMs, WaitForLoadingFast)
}

func clickWithLoading(page playwright.Page, locator playwright.Locator, timeoutMs float64, waitLoading func(playwright.Page) error) error {
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

	return waitLoading(page)
}

func TypeSlowly(page playwright.Page, locator playwright.Locator, text string, timeoutMs float64) error {
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

	if err := page.Keyboard().Press("Control+A"); err != nil {
		return err
	}
	if err := page.Keyboard().Press("Backspace"); err != nil {
		return err
	}

	delay := 50.0
	for _, ch := range text {
		if err := page.Keyboard().Type(string(ch), playwright.KeyboardTypeOptions{
			Delay: playwright.Float(delay),
		}); err != nil {
			return err
		}
	}

	return nil
}

func TypeSlowlyWithLoading(page playwright.Page, locator playwright.Locator, text string, timeoutMs float64) error {
	return typeSlowlyWithLoading(page, locator, text, timeoutMs, WaitForLoading)
}

func TypeSlowlyWithLoadingFast(page playwright.Page, locator playwright.Locator, text string, timeoutMs float64) error {
	return typeSlowlyWithLoading(page, locator, text, timeoutMs, WaitForLoadingFast)
}

func typeSlowlyWithLoading(page playwright.Page, locator playwright.Locator, text string, timeoutMs float64, waitLoading func(playwright.Page) error) error {
	if err := TypeSlowly(page, locator, text, timeoutMs); err != nil {
		return err
	}
	return waitLoading(page)
}
