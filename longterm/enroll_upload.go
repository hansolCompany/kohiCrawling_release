package longterm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

const (
	enrollUploadMenuItemSelector      = `[id="mainframe.VFrameSet.HFrameSet.frameLeft.form.grd_leftMenu.body.gridrow_6.cell_6_0.celltreeitem.treeitemtext"]`
	enrollUploadExcelButtonID         = "mainframe.VFrameSet.HFrameSet.VFrameSetSub.framesetWork.winNPA02310200.form._div_bizFrameMain.form.btn_excelUp:icontext"
	enrollUploadFileNameKeyword       = "보수교육 수강내역 등록"
	enrollUploadFileChooserTimeout    = 60 * time.Second
	enrollUploadButtonClickInterval   = 1 * time.Second
)

func enrollUploadMenuItem(page playwright.Page) playwright.Locator {
	return page.Locator(enrollUploadMenuItemSelector)
}

func enrollUploadExcelButton(page playwright.Page) playwright.Locator {
	return locatorByID(page, enrollUploadExcelButtonID)
}

func clickEnrollUploadMenu(page playwright.Page) error {
	menu := enrollUploadMenuItem(page)
	if err := ClickWithLoading(page, menu, 30000); err != nil {
		return fmt.Errorf("보수교육 수강내역 업로드 메뉴 클릭 실패: %w", err)
	}

	fmt.Println("보수교육 수강내역 업로드 메뉴 클릭 완료")
	return nil
}

func openEnrollUploadFileChooser(page playwright.Page) (playwright.FileChooser, error) {
	button := enrollUploadExcelButton(page)
	if err := button.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		return nil, fmt.Errorf("엑셀 업로드 버튼 대기 실패: %w", err)
	}

	chooserCh := make(chan playwright.FileChooser, 1)
	page.OnFileChooser(func(fileChooser playwright.FileChooser) {
		select {
		case chooserCh <- fileChooser:
		default:
		}
	})

	deadline := time.Now().Add(enrollUploadFileChooserTimeout)
	attempts := 0

	for time.Now().Before(deadline) {
		attempts++
		if err := button.First().Click(); err != nil {
			return nil, fmt.Errorf("엑셀 업로드 버튼 클릭 실패: %w", err)
		}
		fmt.Printf("엑셀 업로드 버튼 클릭 시도: %d\n", attempts)

		select {
		case fileChooser := <-chooserCh:
			fmt.Println("파일 선택 창 열림")
			return fileChooser, nil
		case <-time.After(enrollUploadButtonClickInterval):
		}
	}

	return nil, fmt.Errorf("파일 선택 창 대기 실패: %d초 내에 열리지 않았습니다", int(enrollUploadFileChooserTimeout.Seconds()))
}

func submitEnrollUploadFile(fileChooser playwright.FileChooser, filePath string) error {
	if err := fileChooser.SetFiles(filePath); err != nil {
		return fmt.Errorf("파일 선택 실패: %w", err)
	}

	fmt.Printf("엑셀 파일 선택 완료: %s\n", filePath)
	return nil
}

func resolveEnrollUploadFile() (string, error) {
	desktopDir, err := userDesktopDir()
	if err != nil {
		return "", err
	}

	filePath, err := findFileWithNameKeyword(enrollUploadFileNameKeyword, desktopDir)
	if err != nil {
		return "", fmt.Errorf("바탕화면에서 파일명에 '%s'을(를) 포함한 파일을 찾을 수 없습니다", enrollUploadFileNameKeyword)
	}

	return filePath, nil
}

func userDesktopDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("사용자 홈 디렉터리 확인 실패: %w", err)
	}

	candidates := []string{
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "바탕 화면"),
		filepath.Join(home, "OneDrive", "Desktop"),
		filepath.Join(home, "OneDrive", "바탕 화면"),
	}

	for _, dir := range candidates {
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("바탕화면 폴더를 찾을 수 없습니다")
}

func findFileWithNameKeyword(keyword string, dirs ...string) (string, error) {
	var (
		bestPath string
		bestTime time.Time
		found    bool
	)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.Contains(entry.Name(), keyword) {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			candidate := filepath.Join(dir, entry.Name())
			if !found || info.ModTime().After(bestTime) {
				bestPath = candidate
				bestTime = info.ModTime()
				found = true
			}
		}
	}

	if !found {
		return "", fmt.Errorf("keyword %q not found", keyword)
	}
	return bestPath, nil
}
