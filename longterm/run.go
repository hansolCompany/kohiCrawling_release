package longterm

import (
	"fmt"

	"kohiCrawling/internal/browser"

	"github.com/playwright-community/playwright-go"
)

func Run() error {
	err := runWithSession(DefaultAuth, loginTypeRefresherEducation, "longterm/login", func(page playwright.Page) error {
		fmt.Println("longterm 로그인 완료 — 후속 작업은 API 엔드포인트를 사용하세요")
		return nil
	})
	if err != nil {
		return err
	}

	sharedBrowser.mu.Lock()
	page := sharedBrowser.page
	sharedBrowser.mu.Unlock()
	if page != nil {
		browser.WaitForBrowserClose(page)
	}
	return nil
}

func RunSchedule(auth Auth, items []ScheduleRequest) error {
	return runWithSession(auth, loginTypeRefresherEducation, "longterm/schedule", func(page playwright.Page) error {
		return runSchedule(page, items)
	})
}

func RunEnrollLoadEdu(auth Auth, educationYear string) (*EnrollLoadResult, error) {
	return FetchEnrollLoadEdu(auth, educationYear)
}

func RunEnrollLoadLongterm(auth Auth, educationYear string) (*EnrollLoadResult, error) {
	return FetchEnrollLoadLongterm(auth, educationYear)
}

func FetchEnrollLoadEdu(auth Auth, educationYear string) (*EnrollLoadResult, error) {
	return fetchEnrollLoad(auth, educationYear, loginTypeRefresherEducation, selectRftrObjtrCpetList, "longterm/enrollLoad/edu")
}

func FetchEnrollLoadLongterm(auth Auth, educationYear string) (*EnrollLoadResult, error) {
	return fetchEnrollLoad(auth, educationYear, loginTypeLongtermCareInstitution, selectRftrObjtrList, "longterm/enrollLoad/longterm")
}

func fetchEnrollLoad(auth Auth, educationYear string, loginType string, queryFn func(playwright.Page, string, string) (*EnrollLoadResult, error), taskName string) (*EnrollLoadResult, error) {
	var result *EnrollLoadResult
	err := runWithSession(auth, loginType, taskName, func(page playwright.Page) error {
		var queryErr error
		result, queryErr = queryFn(page, auth.InstitutionCode, educationYear)
		return queryErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func RunEnrollUpload(auth Auth) error {
	return runWithSession(auth, loginTypeRefresherEducation, "longterm/enrollUpload", func(page playwright.Page) error {
		return runEnrollUpload(page)
	})
}

func RunAnnualPlan(auth Auth, educationYear string) error {
	return runWithSession(auth, loginTypeRefresherEducation, "longterm/annualPlan", func(page playwright.Page) error {
		return runAnnualPlan(page, educationYear)
	})
}
