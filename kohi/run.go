package kohi

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"kohiCrawling/internal/browser"
	"kohiCrawling/internal/credentials"

	"github.com/playwright-community/playwright-go"
)

const (
	KohiLoginURL        = "https://edu.kohi.or.kr/in/index.do"
	KohiLearningListURL = "https://edu.kohi.or.kr/in/asp/ac/aca/BD_pfa0010l.do"

	satisfactionAlertKeyword = "과정만족도"
)

var requiredCourseKeywords = []string{"노인인권", "노인학대", "긴급복지", "직장내장애인"}

func loginKohi(page playwright.Page, creds credentials.Credentials) error {
	idInput := page.Locator("#mberId_user")
	if err := idInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("로그인 폼을 찾을 수 없음: %w", err)
	}

	if err := idInput.Fill(creds.UserID); err != nil {
		return fmt.Errorf("아이디 입력 실패: %w", err)
	}

	if err := page.Locator("#user_pw").Fill(creds.Password); err != nil {
		return fmt.Errorf("비밀번호 입력 실패: %w", err)
	}

	browser.ClearDialogMessage()

	_, err := page.ExpectNavigation(func() error {
		return browser.ClickVisible(page.Locator("#btn_login"), 15000)
	}, playwright.PageExpectNavigationOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		if msg := browser.LastDialogMessage(); msg != "" {
			return fmt.Errorf("로그인 실패: %s", msg)
		}
		return fmt.Errorf("로그인 실패: %w", err)
	}

	browser.WaitAfterAction()
	fmt.Println("로그인 성공")
	return nil
}

func courseApplyMenu(page playwright.Page) playwright.Locator {
	menu := page.Locator(".comm-menu-link a").Filter(playwright.LocatorFilterOptions{
		HasText: regexp.MustCompile(`과정신청`),
	})
	if count, _ := menu.Count(); count > 0 {
		return menu
	}

	return page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name: regexp.MustCompile(`과정신청`),
	})
}

func openCourseApplyPage(page playwright.Page) error {
	menu := courseApplyMenu(page)

	_, err := page.ExpectNavigation(func() error {
		return browser.ClickVisible(menu, 15000)
	}, playwright.PageExpectNavigationOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		if clickErr := browser.ClickVisible(menu, 15000); clickErr != nil {
			return fmt.Errorf("과정신청 메뉴 클릭 실패: %w", clickErr)
		}
	}

	if _, err := browser.WaitForPageReady(page); err != nil {
		return fmt.Errorf("과정신청 페이지 로딩 대기 실패: %w", err)
	}

	fmt.Printf("과정신청 페이지 이동 완료: %s\n", page.URL())
	return nil
}

func applyListItems(page playwright.Page) playwright.Locator {
	return page.Locator("#apply_list > ul > li")
}

func countVisibleApplyItems(items playwright.Locator) (int, error) {
	total, err := items.Count()
	if err != nil {
		return 0, err
	}

	visible := 0
	for i := 0; i < total; i++ {
		ok, err := items.Nth(i).IsVisible()
		if err != nil {
			continue
		}
		if ok {
			visible++
		}
	}
	return visible, nil
}

func courseItemTitle(item playwright.Locator) (string, error) {
	text, err := item.Locator("div.tit strong").First().TextContent()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func applyPagingLinks(page playwright.Page) playwright.Locator {
	return page.Locator("#potalPaging a")
}

func getApplyPageNumbers(page playwright.Page) ([]int, error) {
	links := applyPagingLinks(page)
	count, err := links.Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return []int{1}, nil
	}

	seen := make(map[int]struct{})
	var pageNums []int
	for i := 0; i < count; i++ {
		text, err := links.Nth(i).InnerText()
		if err != nil {
			continue
		}
		num, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			continue
		}
		if _, ok := seen[num]; ok {
			continue
		}
		seen[num] = struct{}{}
		pageNums = append(pageNums, num)
	}

	sort.Ints(pageNums)
	if len(pageNums) == 0 {
		return []int{1}, nil
	}
	return pageNums, nil
}

func goToApplyListPage(page playwright.Page, pageNum int) error {
	links := applyPagingLinks(page)
	count, err := links.Count()
	if err != nil {
		return err
	}
	if count == 0 {
		if pageNum == 1 {
			return nil
		}
		return fmt.Errorf("페이지 %d 링크 없음", pageNum)
	}

	target := strconv.Itoa(pageNum)
	for i := 0; i < count; i++ {
		link := links.Nth(i)
		text, err := link.InnerText()
		if err != nil {
			continue
		}
		if strings.TrimSpace(text) != target {
			continue
		}

		if err := browser.ClickVisible(link, 10000); err != nil {
			return fmt.Errorf("페이지 %d 이동 실패: %w", pageNum, err)
		}
		if _, err := browser.WaitForPageReady(page); err != nil {
			return fmt.Errorf("페이지 %d 로딩 대기 실패: %w", pageNum, err)
		}
		return nil
	}

	return fmt.Errorf("페이지 %d 링크 없음", pageNum)
}

func logApplyPageCourses(page playwright.Page, keyword string, pageNum int) {
	items := applyListItems(page)
	total, err := items.Count()
	if err != nil {
		fmt.Printf("[검색] 페이지 %d — 목록 개수 확인 실패: %v\n", pageNum, err)
		return
	}

	visible, _ := countVisibleApplyItems(items)
	fmt.Printf("[검색] 키워드='%s' | 페이지 %d | 표시 %d개 (DOM li %d개)\n", keyword, pageNum, visible, total)

	idx := 0
	for i := 0; i < total; i++ {
		item := items.Nth(i)
		ok, err := item.IsVisible()
		if err != nil || !ok {
			continue
		}

		idx++
		title, err := courseItemTitle(item)
		if err != nil {
			fmt.Printf("  [%d] (강의명 읽기 실패: %v)\n", idx, err)
			continue
		}
		if title == "" {
			fmt.Printf("  [%d] (강의명 없음)\n", idx)
			continue
		}

		status, _ := courseApplyButtonStatus(item)
		if status == "" {
			status = "(버튼 없음)"
		}

		marker := " "
		if strings.Contains(title, keyword) {
			marker = "★"
		}
		fmt.Printf("  [%d] %s %s | 상태: %s\n", idx, marker, title, status)
	}
}

func findCourseItemOnCurrentPage(page playwright.Page, keyword string) (playwright.Locator, string, bool, error) {
	items := applyListItems(page)
	if err := items.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return nil, "", false, fmt.Errorf("과정 목록을 찾을 수 없음: %w", err)
	}

	count, err := items.Count()
	if err != nil {
		return nil, "", false, err
	}

	for i := 0; i < count; i++ {
		item := items.Nth(i)
		ok, err := item.IsVisible()
		if err != nil || !ok {
			continue
		}

		title, err := courseItemTitle(item)
		if err != nil || title == "" {
			continue
		}

		if strings.Contains(title, keyword) {
			return item, title, true, nil
		}
	}

	return nil, "", false, nil
}

func findCourseItemByKeyword(page playwright.Page, keyword string) (playwright.Locator, string, error) {
	pageNums, err := getApplyPageNumbers(page)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("[검색] '%s' 강의 탐색 시작 — 대상 페이지: %v\n", keyword, pageNums)

	for _, pageNum := range pageNums {
		fmt.Printf("[검색] 페이지 %d/%d 이동...\n", pageNum, pageNums[len(pageNums)-1])

		if err := goToApplyListPage(page, pageNum); err != nil {
			return nil, "", err
		}

		logApplyPageCourses(page, keyword, pageNum)

		item, title, found, err := findCourseItemOnCurrentPage(page, keyword)
		if err != nil {
			return nil, "", err
		}
		if found {
			status, _ := courseApplyButtonStatus(item)
			fmt.Printf("[검색] '%s' 매칭 — '%s' (상태: %s)\n", keyword, title, status)
			return item, title, nil
		}

		fmt.Printf("[검색] 페이지 %d — '%s' 강의 없음, 다음 페이지\n", pageNum, keyword)
	}

	fmt.Printf("[검색] '%s' 강의를 모든 페이지(%v)에서 찾지 못함\n", keyword, pageNums)
	return nil, "", fmt.Errorf("'%s' 강의를 찾을 수 없음", keyword)
}

func courseApplyButtonStatus(item playwright.Locator) (string, error) {
	applyBtn := item.Locator("span.status a").First()
	btnCount, err := item.Locator("span.status a").Count()
	if err != nil {
		return "", err
	}
	if btnCount == 0 {
		return "", nil
	}

	text, err := applyBtn.TextContent()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func confirmCourseApplyAlert() {
	msg := waitForDialogMessage(3 * time.Second)
	if msg != "" {
		fmt.Printf("알림 확인: %s\n", msg)
	}
	browser.ClearDialogMessage()
}

func clickCourseApplyButton(item playwright.Locator, title string) error {
	applyBtn := item.Locator("span.status a")
	browser.ClearDialogMessage()
	if err := browser.ClickVisible(applyBtn, 15000); err != nil {
		return fmt.Errorf("수강신청 버튼 클릭 실패 (%s): %w", title, err)
	}
	confirmCourseApplyAlert()
	return nil
}

func finalCourseApplyButton(page playwright.Page) playwright.Locator {
	return page.Locator(`div.btn_wrap a.btn_normal[data-btn="apply_class2"]`)
}

func clickFinalCourseApplyButton(page playwright.Page, title string) error {
	applyBtn := finalCourseApplyButton(page)
	browser.ClearDialogMessage()
	if err := browser.ClickVisible(applyBtn, 15000); err != nil {
		return fmt.Errorf("최종 수강신청 버튼 클릭 실패 (%s): %w", title, err)
	}
	confirmCourseApplyAlert()
	return nil
}

func clickApplyDoneConfirmButton(page playwright.Page, title string) error {
	modal := page.Locator(".ly_pop").Last()
	confirmBtn := modal.GetByRole(*playwright.AriaRoleButton, playwright.LocatorGetByRoleOptions{
		Name:  "확인",
		Exact: playwright.Bool(true),
	})
	confirmLink := modal.GetByRole(*playwright.AriaRoleLink, playwright.LocatorGetByRoleOptions{
		Name:  "확인",
		Exact: playwright.Bool(true),
	})

	btn := page.GetByRole(*playwright.AriaRoleButton, playwright.PageGetByRoleOptions{
		Name:  "확인",
		Exact: playwright.Bool(true),
	})
	link := page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name:  "확인",
		Exact: playwright.Bool(true),
	})

	if err := browser.ClickVisible(confirmBtn.Or(confirmLink).Or(btn).Or(link), 10000); err != nil {
		return fmt.Errorf("확인 버튼 클릭 실패 (%s): %w", title, err)
	}

	fmt.Printf("확인 버튼 클릭 완료: %s\n", title)
	return nil
}

func submitCourseRegistrationModal(page playwright.Page) error {
	modal := page.Locator("#class_opt_agree_pop").Last()
	if err := modal.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("수강신청 모달 대기 실패: %w", err)
	}

	agreeLabel := modal.Locator(`label[for="agreeYes0"]`)
	if err := browser.ClickVisible(agreeLabel, 10000); err != nil {
		return fmt.Errorf("동의 항목 클릭 실패: %w", err)
	}

	agreeLabel2 := modal.Locator(`label[for="agreeNo1"]`)
	if err := browser.ClickVisible(agreeLabel2, 10000); err != nil {
		return fmt.Errorf("동의 항목2 클릭 실패: %w", err)
	}

	registerBtn := modal.GetByRole(*playwright.AriaRoleButton, playwright.LocatorGetByRoleOptions{
		Name:  "등록",
		Exact: playwright.Bool(true),
	})
	link := modal.GetByRole(*playwright.AriaRoleLink, playwright.LocatorGetByRoleOptions{
		Name:  "등록",
		Exact: playwright.Bool(true),
	})
	if err := browser.ClickVisible(registerBtn.Or(link), 10000); err != nil {
		return fmt.Errorf("등록 버튼 클릭 실패: %w", err)
	}

	browser.WaitAfterAction()
	return nil
}

func registerCourseByKeyword(page playwright.Page, keyword string) error {
	item, title, err := findCourseItemByKeyword(page, keyword)
	if err != nil {
		return err
	}

	status, err := courseApplyButtonStatus(item)
	if err != nil {
		return err
	}
	if status == "" {
		fmt.Printf("상태 버튼 없음 — 스킵: %s\n", title)
		return nil
	}
	if strings.Contains(status, "수강완료") {
		fmt.Printf("이미 수강완료 — 스킵: %s\n", title)
		return nil
	}
	if !strings.Contains(status, "수강신청") {
		fmt.Printf("수강신청 불가 (%s) — 스킵: %s\n", status, title)
		return nil
	}

	fmt.Printf("수강신청 시도: %s\n", title)

	if err := clickCourseApplyButton(item, title); err != nil {
		return err
	}

	if err := submitCourseRegistrationModal(page); err != nil {
		return fmt.Errorf("수강신청 모달 처리 실패 (%s): %w", title, err)
	}

	fmt.Printf("수강신청 최종 확인: %s\n", title)
	if err := clickFinalCourseApplyButton(page, title); err != nil {
		return err
	}

	if err := clickApplyDoneConfirmButton(page, title); err != nil {
		return err
	}

	if _, err := browser.WaitForPageReady(page); err != nil {
		return fmt.Errorf("수강신청 완료 후 로딩 대기 실패 (%s): %w", title, err)
	}

	fmt.Printf("수강신청 완료: %s\n", title)
	return nil
}

func registerRequiredCourses(page playwright.Page) error {
	if err := openCourseApplyPage(page); err != nil {
		return err
	}

	for _, keyword := range requiredCourseKeywords {
		if err := registerCourseByKeyword(page, keyword); err != nil {
			return err
		}
	}

	fmt.Println("필수 강의 수강신청 완료")
	return nil
}

func registerCoursesByNames(page playwright.Page, courseNames []string) error {
	if len(courseNames) == 0 {
		return nil
	}

	if err := openCourseApplyPage(page); err != nil {
		return err
	}

	for _, courseName := range courseNames {
		if err := registerCourseByKeyword(page, courseName); err != nil {
			return err
		}
		fmt.Printf("요청 강의 수강신청 처리 완료: %s\n", courseName)
	}

	return nil
}

func learningListMenu(page playwright.Page) playwright.Locator {
	menu := page.Locator(".comm-menu-link a[href*='BD_pfa0010l.do']")
	if count, _ := menu.Count(); count > 0 {
		return menu
	}

	return page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name: regexp.MustCompile(`학습하기\s*/\s*수료증`),
	})
}

func openLearningListPage(page playwright.Page) error {
	menu := learningListMenu(page)

	_, err := page.ExpectNavigation(func() error {
		return browser.ClickVisible(menu, 15000)
	}, playwright.PageExpectNavigationOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		if clickErr := browser.ClickVisible(menu, 15000); clickErr != nil {
			return fmt.Errorf("학습하기/수료증 메뉴 클릭 실패: %w", clickErr)
		}
		if urlErr := page.WaitForURL(regexp.MustCompile(`BD_pfa0010l\.do`), playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(10000),
		}); urlErr != nil {
			return fmt.Errorf("학습하기/수료증 페이지 이동 실패: %w", urlErr)
		}
	}

	if _, err := browser.WaitForPageReady(page); err != nil {
		return fmt.Errorf("학습하기/수료증 페이지 로딩 대기 실패: %w", err)
	}

	fmt.Printf("학습하기/수료증 페이지 이동 완료: %s\n", page.URL())
	return nil
}

func courseTableRows(page playwright.Page) playwright.Locator {
	return page.Locator("table.normal_table.qna tbody tr")
}

func rowCourseStudyButton(row playwright.Locator) playwright.Locator {
	cell := row.Locator("td").Nth(8)

	button := cell.GetByRole(*playwright.AriaRoleButton, playwright.LocatorGetByRoleOptions{
		Name:  "학습하기",
		Exact: playwright.Bool(true),
	})
	link := cell.GetByRole(*playwright.AriaRoleLink, playwright.LocatorGetByRoleOptions{
		Name:  "학습하기",
		Exact: playwright.Bool(true),
	})

	return button.Or(link)
}

func isSatisfactionRequiredAlert() bool {
	return strings.Contains(browser.LastDialogMessage(), satisfactionAlertKeyword)
}

func waitForDialogMessage(timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if msg := browser.LastDialogMessage(); msg != "" {
			return msg
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ""
}

func clickCourseStudyButtonAtRow(page playwright.Page, rowIndex int) error {
	rows := courseTableRows(page)
	row := rows.Nth(rowIndex)
	studyBtn := rowCourseStudyButton(row)

	_, err := page.ExpectNavigation(func() error {
		return browser.ClickVisible(studyBtn, 15000)
	}, playwright.PageExpectNavigationOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		if clickErr := browser.ClickVisible(studyBtn, 15000); clickErr != nil {
			return fmt.Errorf("강의 학습하기 버튼 클릭 실패: %w", clickErr)
		}
	}

	if _, err := browser.WaitForPageReady(page); err != nil {
		return fmt.Errorf("강의 학습하기 클릭 후 로딩 대기 실패: %w", err)
	}

	fmt.Printf("강의 %d번째 학습하기 버튼 클릭 완료\n", rowIndex+1)
	return nil
}

func selectAvailableCourse(page playwright.Page) error {
	if err := openLearningListPage(page); err != nil {
		return err
	}

	rows := courseTableRows(page)
	if err := rows.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("강의 목록을 찾을 수 없음: %w", err)
	}

	count, err := rows.Count()
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		row := rows.Nth(i)
		studyBtn := rowCourseStudyButton(row)
		btnCount, err := studyBtn.Count()
		if err != nil {
			return err
		}
		if btnCount == 0 {
			continue
		}

		fmt.Printf("강의 %d/%d 학습하기 시도...\n", i+1, count)

		browser.ClearDialogMessage()
		if err := clickCourseStudyButtonAtRow(page, i); err != nil {
			return err
		}
		waitForDialogMessage(2 * time.Second)
		if isSatisfactionRequiredAlert() {
			fmt.Println("과정만족도 미참여 — 학습하기/수료증 화면으로 돌아갑니다")
			browser.ClearDialogMessage()
			if err := openLearningListPage(page); err != nil {
				return err
			}
			continue
		}

		browser.ClearDialogMessage()
		if err := clickAgreeButton(page); err != nil {
			return err
		}
		waitForDialogMessage(2 * time.Second)
		if isSatisfactionRequiredAlert() {
			fmt.Println("과정만족도 미참여 — 학습하기/수료증 화면으로 돌아갑니다")
			browser.ClearDialogMessage()
			if err := openLearningListPage(page); err != nil {
				return err
			}
			continue
		}

		browser.ClearDialogMessage()
		fmt.Printf("강의 %d 선택 완료\n", i+1)
		return nil
	}

	return fmt.Errorf("과정만족도 제한 없이 학습 가능한 강의를 찾지 못했습니다")
}

func findCourseRowIndexByName(page playwright.Page, courseName string) (int, error) {
	rows := courseTableRows(page)
	if err := rows.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return -1, fmt.Errorf("강의 목록을 찾을 수 없음: %w", err)
	}

	count, err := rows.Count()
	if err != nil {
		return -1, err
	}

	for i := 0; i < count; i++ {
		text, err := rows.Nth(i).InnerText()
		if err != nil {
			continue
		}
		if strings.Contains(text, courseName) {
			return i, nil
		}
	}

	return -1, fmt.Errorf("'%s' 강의를 찾을 수 없음", courseName)
}

func trySelectCourseRow(page playwright.Page, rowIndex int) error {
	rows := courseTableRows(page)
	row := rows.Nth(rowIndex)
	studyBtn := rowCourseStudyButton(row)
	btnCount, err := studyBtn.Count()
	if err != nil {
		return err
	}
	if btnCount == 0 {
		return fmt.Errorf("강의 %d번째 행에 학습하기 버튼이 없습니다", rowIndex+1)
	}

	fmt.Printf("강의 %d 학습하기 시도...\n", rowIndex+1)

	browser.ClearDialogMessage()
	if err := clickCourseStudyButtonAtRow(page, rowIndex); err != nil {
		return err
	}
	waitForDialogMessage(2 * time.Second)
	if isSatisfactionRequiredAlert() {
		return fmt.Errorf("과정만족도 미참여로 학습할 수 없습니다")
	}

	browser.ClearDialogMessage()
	if err := clickAgreeButton(page); err != nil {
		return err
	}
	waitForDialogMessage(2 * time.Second)
	if isSatisfactionRequiredAlert() {
		return fmt.Errorf("과정만족도 미참여로 학습할 수 없습니다")
	}

	browser.ClearDialogMessage()
	fmt.Printf("강의 %d 선택 완료\n", rowIndex+1)
	return nil
}

func selectCourseByName(page playwright.Page, courseName string) error {
	if err := openLearningListPage(page); err != nil {
		return err
	}

	rowIndex, err := findCourseRowIndexByName(page, courseName)
	if err != nil {
		return err
	}

	fmt.Printf("강의명 '%s' 검색 완료 (행 %d)\n", courseName, rowIndex+1)
	return trySelectCourseRow(page, rowIndex)
}

func rowStudyButton(row playwright.Locator) playwright.Locator {
	button := row.GetByRole(*playwright.AriaRoleButton, playwright.LocatorGetByRoleOptions{
		Name:  "학습하기",
		Exact: playwright.Bool(true),
	})
	link := row.GetByRole(*playwright.AriaRoleLink, playwright.LocatorGetByRoleOptions{
		Name:  "학습하기",
		Exact: playwright.Bool(true),
	})

	return button.Or(link)
}

func findIncompleteVideoRow(page playwright.Page) (playwright.Locator, error) {
	rows := page.Locator("tr.openStudyTr")
	if err := rows.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return nil, fmt.Errorf("영상 목록을 찾을 수 없음: %w", err)
	}

	count, err := rows.Count()
	if err != nil {
		return nil, err
	}

	for i := 0; i < count; i++ {
		row := rows.Nth(i)
		progress, err := row.Locator("td").Nth(5).InnerText()
		if err != nil {
			return nil, fmt.Errorf("진도율 확인 실패: %w", err)
		}

		progress = strings.TrimSpace(progress)
		if !strings.Contains(progress, "100%") {
			fmt.Printf("미완료 영상 발견 (진도율: %s)\n", progress)
			return row, nil
		}
	}

	return nil, nil
}

func openIncompleteVideoStudyWindow(page playwright.Page) (playwright.Page, bool, error) {
	row, err := findIncompleteVideoRow(page)
	if err != nil {
		return nil, false, err
	}
	if row == nil {
		return nil, false, nil
	}

	studyBtn := rowStudyButton(row)
	popup, err := page.ExpectPopup(func() error {
		return browser.ClickVisible(studyBtn, 15000)
	}, playwright.PageExpectPopupOptions{
		Timeout: playwright.Float(30000),
	})
	if err != nil {
		return nil, false, fmt.Errorf("미완료 영상 학습하기 클릭 실패: %w", err)
	}

	if err := popup.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	}); err != nil {
		return nil, false, fmt.Errorf("학습 새 창 로딩 실패: %w", err)
	}

	if _, err := browser.WaitForPageReady(popup); err != nil {
		return nil, false, fmt.Errorf("학습 새 창 로딩 대기 실패: %w", err)
	}

	fmt.Println("미완료 영상 학습 새 창 열림")
	return popup, true, nil
}

func agreeButton(page playwright.Page) playwright.Locator {
	button := page.GetByRole(*playwright.AriaRoleButton, playwright.PageGetByRoleOptions{
		Name: regexp.MustCompile(`^동의$|^동의합니다$|^동의하기$`),
	})
	link := page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name: regexp.MustCompile(`^동의$|^동의합니다$|^동의하기$`),
	})

	return button.Or(link)
}

func clickAgreeButton(page playwright.Page) error {
	agreeBtn := agreeButton(page)

	_, err := page.ExpectNavigation(func() error {
		return browser.ClickVisible(agreeBtn, 15000)
	}, playwright.PageExpectNavigationOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		if clickErr := browser.ClickVisible(agreeBtn, 15000); clickErr != nil {
			return fmt.Errorf("동의 버튼 클릭 실패: %w", clickErr)
		}
	}

	if _, err := browser.WaitForPageReady(page); err != nil {
		return fmt.Errorf("동의 후 로딩 대기 실패: %w", err)
	}

	fmt.Println("동의 버튼 클릭 완료")
	return nil
}

func studyFinishButton(page playwright.Page) playwright.Locator {
	button := page.GetByRole(*playwright.AriaRoleButton, playwright.PageGetByRoleOptions{
		Name: regexp.MustCompile(`학습종료`),
	})
	link := page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
		Name: regexp.MustCompile(`학습종료`),
	})

	return button.Or(link)
}

func startStudyVideo(studyPage playwright.Page) error {
	if err := studyPage.BringToFront(); err != nil {
		return fmt.Errorf("학습 창 포커스 실패: %w", err)
	}

	playBtn := studyPage.Locator(".plyr__control.plyr__control--overlaid")
	if err := browser.ClickVisible(playBtn, 60000); err != nil {
		return fmt.Errorf("영상 재생 버튼 클릭 실패: %w", err)
	}

	fmt.Println("영상 재생 버튼 클릭 완료")
	return nil
}

func getVideoTime(studyPage playwright.Page) (string, string, error) {
	currentLoc := studyPage.Locator(".plyr__controls .plyr__time--current.plyr__time").First()
	durationLoc := studyPage.Locator(".plyr__controls .plyr__time--duration.plyr__time").First()

	current, err := currentLoc.InnerText()
	if err != nil {
		return "", "", err
	}

	duration, err := durationLoc.InnerText()
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(current), strings.TrimSpace(duration), nil
}

func isVideoEnded(current, duration string) bool {
	return current != "" && duration != "" && current == duration
}

func getPagination(studyPage playwright.Page) (int, int, bool, error) {
	numLoc := studyPage.Locator(".paging_area .num")
	count, err := numLoc.Count()
	if err != nil {
		return 0, 0, false, err
	}
	if count == 0 {
		return 0, 0, false, nil
	}

	text, err := numLoc.First().InnerText()
	if err != nil {
		return 0, 0, false, err
	}

	matches := regexp.MustCompile(`(\d+)\s*/\s*(\d+)`).FindStringSubmatch(text)
	if len(matches) != 3 {
		return 0, 0, false, nil
	}

	current, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, false, err
	}

	total, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, false, err
	}

	return current, total, true, nil
}

func clickFinishStudy(studyPage playwright.Page) error {
	finishBtn := studyFinishButton(studyPage)
	if err := browser.ClickVisible(finishBtn, 10000); err != nil {
		return fmt.Errorf("학습종료 버튼 클릭 실패: %w", err)
	}

	if waitUntilPageClosed(studyPage, 3*time.Second) {
		fmt.Println("학습 종료 및 창 닫힘 확인")
		return nil
	}

	fmt.Println("학습종료 후 창이 닫히지 않아 버튼을 다시 클릭합니다")
	if studyPage.IsClosed() {
		fmt.Println("학습 종료 및 창 닫힘 확인")
		return nil
	}

	if err := browser.ClickVisible(finishBtn, 5000); err != nil {
		return fmt.Errorf("학습종료 버튼 재클릭 실패: %w", err)
	}

	if waitUntilPageClosed(studyPage, 3*time.Second) {
		fmt.Println("학습 종료 및 창 닫힘 확인")
		return nil
	}

	return fmt.Errorf("학습종료 후 창이 닫히지 않음")
}

func waitUntilPageClosed(page playwright.Page, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if page.IsClosed() {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return page.IsClosed()
}

func waitForVideoEnd(studyPage playwright.Page) error {
	for {
		if studyPage.IsClosed() {
			return nil
		}

		current, duration, err := getVideoTime(studyPage)
		if err != nil {
			return fmt.Errorf("영상 시간 확인 실패: %w", err)
		}

		if isVideoEnded(current, duration) {
			fmt.Printf("영상 종료 확인: %s / %s\n", current, duration)
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

func handleAfterVideoEnd(studyPage playwright.Page) (bool, error) {
	if studyPage.IsClosed() {
		return true, nil
	}

	pageNum, totalPages, ok, err := getPagination(studyPage)
	if err != nil {
		return false, err
	}

	if ok && pageNum == totalPages {
		fmt.Printf("마지막 페이지(%d/%d) - 학습종료\n", pageNum, totalPages)
		if err := clickFinishStudy(studyPage); err != nil {
			return false, err
		}
		return true, nil
	}

	if ok {
		fmt.Printf("다음 페이지 이동: %d/%d\n", pageNum, totalPages)
	} else {
		fmt.Println("다음 페이지 이동")
	}

	nextBtn := studyPage.Locator("#paging_next")
	if err := browser.ClickVisible(nextBtn, 10000); err != nil {
		return false, fmt.Errorf("다음 페이지 버튼 클릭 실패: %w", err)
	}

	if _, err := browser.WaitForPageReady(studyPage); err != nil {
		return false, fmt.Errorf("다음 페이지 로딩 대기 실패: %w", err)
	}

	return false, nil
}

func runStudySession(studyPage playwright.Page) error {
	for {
		if studyPage.IsClosed() {
			return nil
		}

		if err := startStudyVideo(studyPage); err != nil {
			return err
		}

		if err := waitForVideoEnd(studyPage); err != nil {
			return err
		}

		finished, err := handleAfterVideoEnd(studyPage)
		if err != nil {
			return err
		}
		if finished || studyPage.IsClosed() {
			return nil
		}
	}
}

func runCourseStudyLoop(page playwright.Page) (playwright.Page, error) {
	for {
		studyPage, hasIncomplete, err := openIncompleteVideoStudyWindow(page)
		if err != nil {
			return page, err
		}
		if !hasIncomplete {
			fmt.Println("강의 내 모든 영상 학습 완료 (진도율 100%)")
			return page, nil
		}

		if err := runStudySession(studyPage); err != nil {
			return page, err
		}

		if _, err := browser.WaitForPageReady(page); err != nil {
			return page, fmt.Errorf("영상 선택 페이지 로딩 대기 실패: %w", err)
		}
	}
}

func goToCourseStudy(page playwright.Page) (playwright.Page, error) {
	if err := selectAvailableCourse(page); err != nil {
		return page, err
	}
	return runCourseStudyLoop(page)
}

func goToCourseStudyByName(page playwright.Page, courseName string) (playwright.Page, error) {
	if err := selectCourseByName(page, courseName); err != nil {
		return page, err
	}
	return runCourseStudyLoop(page)
}

func runWithBrowser(taskName string, fn func(page playwright.Page) error) error {
	if err := browser.EnsurePlaywrightBrowser(); err != nil {
		return err
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright 실행 실패: %w", err)
	}
	defer pw.Stop()

	browserInstance, page, err := browser.LaunchVisibleBrowser(pw)
	if err != nil {
		return fmt.Errorf("브라우저 실행 실패: %w", err)
	}
	defer browserInstance.Close()

	browser.SetupDialogHandler(page.Context())

	_, err = page.Goto(KohiLoginURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return fmt.Errorf("사이트 접속 실패: %w", err)
	}
	browser.WaitAfterAction()

	fmt.Printf("[%s] 접속 완료: %s\n", taskName, KohiLoginURL)
	return fn(page)
}

func Run(creds credentials.Credentials) error {
	return runWithBrowser("kohi", func(page playwright.Page) error {
		if err := loginKohi(page, creds); err != nil {
			return err
		}
		if err := registerRequiredCourses(page); err != nil {
			return err
		}
		studyPage, err := goToCourseStudy(page)
		if err != nil {
			return err
		}
		browser.WaitForBrowserClose(studyPage)
		return nil
	})
}

func RunAutoLearn(opts AutoLearnOptions) error {
	creds := credentials.Credentials{
		UserID:   opts.UserID,
		Password: opts.Password,
	}

	return runWithSession(creds, "kohi/autoLearn", func(page playwright.Page) (playwright.Page, error) {
		fmt.Printf("요양보호사 DB ID: %s\n", opts.CaregiverDbID)
		fmt.Printf("수강 강의 (%d건): %v\n", len(opts.CourseNames), opts.CourseNames)

		if err := registerCoursesByNames(page, opts.CourseNames); err != nil {
			return page, err
		}

		studyPage := page
		for i, courseName := range opts.CourseNames {
			fmt.Printf("강의 학습 시작 [%d/%d]: %s\n", i+1, len(opts.CourseNames), courseName)
			var err error
			studyPage, err = goToCourseStudyByName(page, courseName)
			if err != nil {
				return page, err
			}
		}

		return studyPage, nil
	})
}
