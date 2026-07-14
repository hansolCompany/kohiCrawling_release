package longterm

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"kohiCrawling/internal/browser"

	"github.com/playwright-community/playwright-go"
)

const (
	LoginURL = "https://www.longtermcare.or.kr/npbs/auth/login/loginForm.web?menuId=npe0000002840&rtnUrl=&zoomSize="

	loginTypeLongtermCareInstitution = "A"
	loginTypeRefresherEducation      = "E"
)

var loginTypeLabels = map[string]string{
	loginTypeLongtermCareInstitution: "장기요양기관",
	loginTypeRefresherEducation:      "보수교육기관",
}

func prepareLongtermLogin(page playwright.Page, auth Auth, loginType string) (playwright.Page, error) {
	if _, err := page.Goto(LoginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return page, fmt.Errorf("로그인 페이지 접속 실패: %w", err)
	}

	if _, err := browser.WaitForPageReady(page); err != nil {
		return page, fmt.Errorf("로그인 페이지 로딩 대기 실패: %w", err)
	}

	fmt.Printf("접속 완료: %s\n", LoginURL)

	// if err := checkKeyboardSecurity(page); err != nil {
	// 	return page, err
	// }

	if err := clickOrgLoginTab(page); err != nil {
		return page, err
	}

	if err := selectLoginType(page, loginType); err != nil {
		return page, err
	}

	if err := enterInstitutionCode(page, auth.InstitutionCode); err != nil {
		return page, err
	}

	if err := clickCorporateCertLogin(page); err != nil {
		return page, err
	}

	if err := selectCorporateCertificate(page, auth); err != nil {
		return page, err
	}

	activePage, err := focusNewTabAfterLogin(page, 30*time.Second)
	if err != nil {
		return page, err
	}

	fmt.Println("법인인증서 선택 및 비밀번호 입력 완료")
	return activePage, nil
}

func clickOrgLoginTab(page playwright.Page) error {
	tab := page.Locator("#tab_login_02 .btn-tab")
	if err := tab.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("장기요양 관련기관 로그인 탭을 찾을 수 없음: %w", err)
	}

	if err := browser.ClickVisible(tab, 15000); err != nil {
		return fmt.Errorf("장기요양 관련기관 로그인 탭 클릭 실패: %w", err)
	}

	panel := page.Locator("#panel_login_02")
	if err := panel.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("장기요양 관련기관 로그인 패널 표시 대기 실패: %w", err)
	}

	fmt.Println("장기요양 관련기관 로그인 탭 선택 완료")
	return nil
}

func checkKeyboardSecurity(page playwright.Page) error {
	checkbox := page.Locator("#chkTouchEn")
	if err := checkbox.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("키보드보안 체크박스를 찾을 수 없음: %w", err)
	}

	label := page.Locator("label[for='chkTouchEn']")
	deadline := time.Now().Add(2 * time.Minute)

	for time.Now().Before(deadline) {
		checked, err := checkbox.IsChecked()
		if err != nil {
			return fmt.Errorf("키보드보안 체크박스 상태 확인 실패: %w", err)
		}
		if checked {
			fmt.Println("키보드보안 프로그램 선택 설치 체크 완료")
			return nil
		}

		browser.ClearDialogMessage()

		target := label.First()
		if err := target.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("키보드보안 체크박스 라벨을 찾을 수 없음: %w", err)
		}
		if err := target.Click(); err != nil {
			return fmt.Errorf("키보드보안 프로그램 선택 설치 체크 실패: %w", err)
		}

		if msg := waitForDialogConfirmation(5 * time.Second); msg != "" {
			fmt.Printf("알림 확인: %s\n", msg)
		}

		time.Sleep(3 * time.Second)

		checked, err = checkbox.IsChecked()
		if err != nil {
			return fmt.Errorf("키보드보안 체크박스 재확인 실패: %w", err)
		}
		if checked {
			fmt.Println("키보드보안 프로그램 선택 설치 체크 완료")
			return nil
		}

		fmt.Println("키보드보안 체크박스 미선택 — 재시도...")
	}

	return fmt.Errorf("키보드보안 프로그램 선택 설치 체크박스가 선택되지 않았습니다")
}

func waitForDialogConfirmation(timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if msg := browser.LastDialogMessage(); msg != "" {
			return msg
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ""
}

func selectLoginType(page playwright.Page, loginType string) error {
	label := loginTypeLabels[loginType]
	if label == "" {
		label = loginType
	}

	selectBox := page.Locator("#user_type_dis")
	if err := selectBox.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("로그인 유형 선택을 찾을 수 없음: %w", err)
	}

	if _, err := selectBox.SelectOption(playwright.SelectOptionValues{Values: &[]string{loginType}}); err != nil {
		return fmt.Errorf("%s 선택 실패: %w", label, err)
	}

	selected, err := selectBox.InputValue()
	if err != nil {
		return fmt.Errorf("로그인 유형 선택값 확인 실패: %w", err)
	}
	if selected != loginType {
		return fmt.Errorf("로그인 유형이 %s(%s)으로 설정되지 않았습니다 (현재: %s)", label, loginType, selected)
	}

	fmt.Printf("로그인 유형: %s 선택 완료\n", label)
	return nil
}

func enterInstitutionCode(page playwright.Page, code string) error {
	input := page.Locator("#userNo")
	if err := input.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("장기요양기관기호 입력란을 찾을 수 없음: %w", err)
	}

	if err := input.Fill(code); err != nil {
		return fmt.Errorf("장기요양기관기호 입력 실패: %w", err)
	}

	value, err := input.InputValue()
	if err != nil {
		return fmt.Errorf("장기요양기관기호 입력값 확인 실패: %w", err)
	}
	if value != code {
		return fmt.Errorf("장기요양기관기호가 올바르게 입력되지 않았습니다 (입력값: %s)", value)
	}

	fmt.Printf("장기요양기관기호 입력 완료: %s\n", code)
	return nil
}

func clickCorporateCertLogin(page playwright.Page) error {
	button := page.Locator("#btn_login_A2_A")
	if err := button.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("법인인증서 로그인 버튼을 찾을 수 없음: %w", err)
	}

	text, err := button.InnerText()
	if err != nil {
		return fmt.Errorf("법인인증서 로그인 버튼 텍스트 확인 실패: %w", err)
	}

	if !regexp.MustCompile(`법인\s*인증서`).MatchString(text) {
		return fmt.Errorf("법인인증서 로그인 버튼이 활성화되지 않았습니다 (버튼 텍스트: %s)", text)
	}

	if err := browser.ClickVisible(button, 15000); err != nil {
		return fmt.Errorf("법인인증서 로그인 버튼 클릭 실패: %w", err)
	}

	return nil
}

func selectCorporateCertificate(page playwright.Page, auth Auth) error {
	root, err := waitForCertRoot(page)
	if err != nil {
		return err
	}

	if err := clickCertLocation(page, root); err != nil {
		return err
	}

	if err := selectCertFromTable(root, auth.CertName); err != nil {
		return err
	}

	passwordInput := root.Locator("#xwup_certselect_tek_input1")
	if err := passwordInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("인증서 비밀번호 입력란을 찾을 수 없음: %w", err)
	}

	if err := fillCertPassword(page, passwordInput, auth.CertPassword); err != nil {
		return err
	}

	if err := clickCertOkButton(page, root); err != nil {
		return err
	}

	fmt.Printf("인증서 선택 완료: %s\n", auth.CertName)
	return nil
}

func selectCertFromTable(root playwright.Locator, certName string) error {
	certTable := root.Locator("#xwup_cert_table")
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		rows := certTable.Locator("tr")
		count, err := rows.Count()
		if err == nil && count > 0 {
			for i := 0; i < count; i++ {
				row := rows.Nth(i)
				cells := row.Locator("td")
				cellCount, err := cells.Count()
				if err != nil || cellCount < 2 {
					continue
				}

				text, err := cells.Nth(1).InnerText()
				if err != nil {
					continue
				}
				cellText := strings.TrimSpace(text)
				if !strings.Contains(cellText, certName) {
					continue
				}

				if err := browser.ClickVisible(row, 15000); err != nil {
					return fmt.Errorf("인증서 '%s' 선택 실패: %w", certName, err)
				}
				fmt.Printf("인증서 선택: %s (검색어: %s)\n", cellText, certName)
				return nil
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("인증서 '%s'를 목록에서 찾을 수 없음 (5초 타임아웃)", certName)
}

func fillCertPassword(page playwright.Page, passwordInput playwright.Locator, password string) error {
	_ = page
	const targetSelector = "#xwup_certselect_tek_input1"

	result, err := passwordInput.Evaluate(`(passwordField, password) => {
		if (!passwordField) {
			return null;
		}

		passwordField.removeAttribute('readonly');
		passwordField.readOnly = false;
		passwordField.value = password;

		const inputEvent = new Event('input', { bubbles: true });
		const changeEvent = new Event('change', { bubbles: true });
		const keydownEvent = new KeyboardEvent('keydown', { bubbles: true });
		const keyupEvent = new KeyboardEvent('keyup', { bubbles: true });

		passwordField.dispatchEvent(keydownEvent);
		passwordField.dispatchEvent(inputEvent);
		passwordField.dispatchEvent(changeEvent);
		passwordField.dispatchEvent(keyupEvent);

		return passwordField.value;
	}`, password)
	if err != nil {
		return fmt.Errorf("인증서 비밀번호 스크립트 입력 실패: %w", err)
	}

	if result == nil {
		return fmt.Errorf("인증서 비밀번호 입력란을 찾을 수 없음")
	}

	value, ok := result.(string)
	if !ok {
		return fmt.Errorf("인증서 비밀번호 입력값 확인 실패")
	}

	if value != password {
		return fmt.Errorf("인증서 비밀번호가 올바르게 입력되지 않았습니다")
	}

	browser.WaitAfterAction()
	return nil
}

func maskSecret(value string) string {
	switch len(value) {
	case 0:
		return ""
	case 1:
		return "*"
	case 2:
		return "**"
	case 3:
		return value[:1] + "*" + value[2:]
	default:
		return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
	}
}

func focusNewTabAfterLogin(loginPage playwright.Page, timeout time.Duration) (playwright.Page, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		for _, targetPage := range loginPage.Context().Pages() {
			if targetPage == loginPage || targetPage.IsClosed() {
				continue
			}

			if err := targetPage.BringToFront(); err != nil {
				return nil, fmt.Errorf("새 탭 포커스 실패: %w", err)
			}

			if err := targetPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			}); err != nil {
				return nil, fmt.Errorf("새 탭 로딩 대기 실패: %w", err)
			}

			if err := WaitForLoading(targetPage); err != nil {
				return nil, fmt.Errorf("로그인 후 초기 로딩 대기 실패: %w", err)
			}

			return targetPage, nil
		}

		time.Sleep(200 * time.Millisecond)
	}

	return nil, fmt.Errorf("로그인 후 새 탭을 찾을 수 없음 (%s)", timeout)
}

func clickCertOkButton(page playwright.Page, root playwright.Locator) error {
	okButton, err := waitForCertOkButton(page, root)
	if err != nil {
		return err
	}

	clicked, err := okButton.Evaluate(`(button) => {
		if (!button) {
			return false;
		}

		button.removeAttribute('disabled');
		button.disabled = false;

		button.dispatchEvent(new Event('focus', { bubbles: true }));

		for (const type of ['mousedown', 'mouseup', 'click']) {
			button.dispatchEvent(new MouseEvent(type, {
				bubbles: true,
				cancelable: true,
				view: window,
			}));
		}

		if (typeof button.click === 'function') {
			button.click();
		}
		if (typeof button.onclick === 'function') {
			button.onclick(new MouseEvent('click', { bubbles: true, cancelable: true, view: window }));
		}

		return true;
	}`, nil)
	if err != nil {
		return fmt.Errorf("인증서 확인 버튼 스크립트 클릭 실패: %w", err)
	}

	clickedOK, ok := clicked.(bool)
	if !ok || !clickedOK {
		if err := browser.ClickVisible(okButton, 15000); err != nil {
			return fmt.Errorf("인증서 확인 버튼 클릭 실패: %w", err)
		}
	}

	browser.WaitAfterAction()
	return nil
}

func waitForCertOkButton(page playwright.Page, root playwright.Locator) (playwright.Locator, error) {
	okButton := root.Locator("#xwup_OkButton")
	if count, err := okButton.Count(); err == nil && count > 0 {
		if err := okButton.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		}); err == nil {
			return okButton.First(), nil
		}
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, targetPage := range page.Context().Pages() {
			if loc, ok := findCertElement(targetPage, "#xwup_OkButton"); ok {
				return loc, nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil, fmt.Errorf("인증서 확인 버튼(#xwup_OkButton)을 찾을 수 없음")
}

func clickCertLocation(page playwright.Page, root playwright.Locator) error {
	location := root.Locator("#xwup_location_3")
	count, err := location.Count()
	if err == nil && count > 0 {
		if err := browser.ClickVisible(location, 15000); err != nil {
			return fmt.Errorf("인증서 위치 선택 클릭 실패: %w", err)
		}
		return nil
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		for _, targetPage := range page.Context().Pages() {
			if loc, ok := findCertElement(targetPage, "#xwup_location_3"); ok {
				if err := browser.ClickVisible(loc, 15000); err != nil {
					return fmt.Errorf("인증서 위치 선택 클릭 실패: %w", err)
				}
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("인증서 위치 선택(#xwup_location_3)을 찾을 수 없음")
}

func waitForCertRoot(page playwright.Page) (playwright.Locator, error) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		for _, targetPage := range page.Context().Pages() {
			if root, ok := findCertRoot(targetPage); ok {
				return root, nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil, fmt.Errorf("인증서 선택 창(#xTsign)을 찾을 수 없음")
}

func findCertRoot(page playwright.Page) (playwright.Locator, bool) {
	if root, ok := findCertRootInLocator(page.Locator("body")); ok {
		return root, true
	}

	for _, frame := range page.Frames() {
		if root, ok := findCertRootInLocator(frame.Locator("body")); ok {
			return root, true
		}
	}
	return nil, false
}

func findCertRootInLocator(scope playwright.Locator) (playwright.Locator, bool) {
	return findVisibleElement(scope, "#xTsign")
}

func findCertElement(page playwright.Page, selector string) (playwright.Locator, bool) {
	if loc, ok := findVisibleElement(page.Locator("body"), selector); ok {
		return loc, true
	}

	for _, frame := range page.Frames() {
		if loc, ok := findVisibleElement(frame.Locator("body"), selector); ok {
			return loc, true
		}
	}
	return nil, false
}

func findVisibleElement(scope playwright.Locator, selector string) (playwright.Locator, bool) {
	loc := scope.Locator(selector)
	count, err := loc.Count()
	if err != nil || count == 0 {
		return nil, false
	}

	target := loc.First()
	visible, err := target.IsVisible()
	if err != nil || !visible {
		return nil, false
	}
	return target, true
}
