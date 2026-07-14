package longterm

import (
	"fmt"
	"sync"

	"kohiCrawling/internal/browser"

	"github.com/playwright-community/playwright-go"
)

type browserSession struct {
	mu        sync.Mutex
	pw        *playwright.Playwright
	browser   playwright.Browser
	page      playwright.Page
	auth      Auth
	loginType string
	active    bool
}

var sharedBrowser browserSession

func (a Auth) matches(other Auth) bool {
	return a.InstitutionCode == other.InstitutionCode &&
		a.CertName == other.CertName &&
		a.CertPassword == other.CertPassword
}

func (s *browserSession) sessionMatches(auth Auth, loginType string) bool {
	if !s.active || s.page == nil || browser.IsClosed(s.page) {
		return false
	}
	return s.auth.matches(auth) && s.loginType == loginType
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

func (s *browserSession) launchAndLogin(auth Auth, loginType string) error {
	if err := browser.EnsurePlaywrightBrowser(); err != nil {
		return err
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright 실행 실패: %w", err)
	}

	browserInstance, page, err := browser.LaunchVisibleBrowser(pw)
	if err != nil {
		_ = pw.Stop()
		return err
	}

	browser.SetupDialogHandler(page.Context())

	page, loginErr := prepareLongtermLogin(page, auth, loginType)
	if loginErr != nil {
		_ = browserInstance.Close()
		_ = pw.Stop()
		return loginErr
	}

	s.pw = pw
	s.browser = browserInstance
	s.page = page
	s.auth = auth
	s.loginType = loginType
	s.active = true
	return nil
}

func (s *browserSession) ensure(auth Auth, loginType string) (playwright.Page, error) {
	if s.sessionMatches(auth, loginType) {
		return s.page, nil
	}

	if s.active {
		fmt.Println("계정 또는 로그인 유형 변경 — 브라우저 세션 재시작")
	}
	s.closeLocked()

	if err := s.launchAndLogin(auth, loginType); err != nil {
		return nil, err
	}
	return s.page, nil
}

func (s *browserSession) setPage(page playwright.Page) {
	s.page = page
}

func runWithSession(auth Auth, loginType, taskName string, fn func(page playwright.Page) error) error {
	sharedBrowser.mu.Lock()
	defer sharedBrowser.mu.Unlock()

	page, err := sharedBrowser.ensure(auth, loginType)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] 작업 시작\n", taskName)
	return fn(page)
}
