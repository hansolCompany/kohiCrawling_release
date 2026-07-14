package kohi

import (
	"fmt"
	"sync"

	"kohiCrawling/internal/browser"
	"kohiCrawling/internal/credentials"

	"github.com/playwright-community/playwright-go"
)

type browserSession struct {
	mu     sync.Mutex
	pw     *playwright.Playwright
	browser playwright.Browser
	page   playwright.Page
	creds  credentials.Credentials
	active bool
}

var sharedBrowser browserSession

func credentialsMatch(a, b credentials.Credentials) bool {
	return a.UserID == b.UserID && a.Password == b.Password
}

func (s *browserSession) sessionMatches(creds credentials.Credentials) bool {
	if !s.active || s.page == nil || browser.IsClosed(s.page) {
		return false
	}
	return credentialsMatch(s.creds, creds)
}

func (s *browserSession) closeLocked() {
	if s.browser != nil {
		_ = s.browser.Close()
	}
	if s.pw != nil {
		_ = s.pw.Stop()
	}
	s.pw = nil
	s.browser = nil
	s.page = nil
	s.active = false
}

func (s *browserSession) launchAndLogin(creds credentials.Credentials) (playwright.Page, error) {
	if err := browser.EnsurePlaywrightBrowser(); err != nil {
		return nil, err
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright 실행 실패: %w", err)
	}

	browserInstance, page, err := browser.LaunchVisibleBrowser(pw)
	if err != nil {
		_ = pw.Stop()
		return nil, err
	}

	browser.SetupDialogHandler(page.Context())

	if _, err := page.Goto(KohiLoginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		_ = browserInstance.Close()
		_ = pw.Stop()
		return nil, fmt.Errorf("사이트 접속 실패: %w", err)
	}
	browser.WaitAfterAction()

	if err := loginKohi(page, creds); err != nil {
		_ = browserInstance.Close()
		_ = pw.Stop()
		return nil, err
	}

	s.pw = pw
	s.browser = browserInstance
	s.page = page
	s.creds = creds
	s.active = true
	return page, nil
}

func (s *browserSession) ensure(creds credentials.Credentials) (playwright.Page, error) {
	if s.sessionMatches(creds) {
		return s.page, nil
	}

	if s.active {
		fmt.Println("계정 변경 — 브라우저 세션 재시작")
	}
	s.closeLocked()

	page, err := s.launchAndLogin(creds)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (s *browserSession) setPage(page playwright.Page) {
	s.page = page
}

func runWithSession(creds credentials.Credentials, taskName string, fn func(page playwright.Page) (playwright.Page, error)) error {
	sharedBrowser.mu.Lock()
	defer sharedBrowser.mu.Unlock()

	page, err := sharedBrowser.ensure(creds)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] 작업 시작\n", taskName)
	page, err = fn(page)
	if err != nil {
		return err
	}
	sharedBrowser.setPage(page)
	return nil
}
