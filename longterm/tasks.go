package longterm

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func runSchedule(page playwright.Page, items []ScheduleRequest) error {
	fmt.Println("보수교육 일정관리 작업")
	fmt.Printf("  일정 건수: %d\n", len(items))

	for i, req := range items {
		fmt.Printf("  [%d] 교육일자: %s\n", i+1, req.EducationDate)
		fmt.Printf("  [%d] 표준영역: %s\n", i+1, req.StandardArea)
		fmt.Printf("  [%d] 과목명: %s\n", i+1, req.CourseName)
		fmt.Printf("  [%d] 시작시간: %s\n", i+1, req.CourseStartTime)
		fmt.Printf("  [%d] 종료시간: %s\n", i+1, req.CourseEndTime)
		fmt.Printf("  [%d] 강사명: %s\n", i+1, req.InstructorName)
		fmt.Printf("  [%d] 정원: %d\n", i+1, req.Capacity)
		if req.Note != "" {
			fmt.Printf("  [%d] 비고: %s\n", i+1, req.Note)
		}
	}
	fmt.Printf("  현재 URL: %s\n", page.URL())

	if err := clickScheduleMenu(page); err != nil {
		return err
	}

	for i, req := range items {
		fmt.Printf("일정 [%d/%d] 처리 중...\n", i+1, len(items))
		if err := prepareScheduleRow(page, req); err != nil {
			return err
		}
	}

	return nil
}

func runEnrollUpload(page playwright.Page) error {
	fmt.Println("보수교육 수강내역 업로드 작업")
	fmt.Printf("  현재 URL: %s\n", page.URL())

	if err := clickEnrollUploadMenu(page); err != nil {
		return err
	}

	fileChooser, err := openEnrollUploadFileChooser(page)
	if err != nil {
		return err
	}

	filePath, err := resolveEnrollUploadFile()
	if err != nil {
		return err
	}
	fmt.Printf("  엑셀 파일: %s\n", filePath)

	if err := submitEnrollUploadFile(fileChooser, filePath); err != nil {
		return err
	}

	return WaitForLoading(page)
}

func runAnnualPlan(page playwright.Page, educationYear string) error {
	fmt.Println("연간 보수교육 사업계획 작업")
	fmt.Printf("  교육년도: %s\n", educationYear)
	fmt.Printf("  현재 URL: %s\n", page.URL())

	// TODO: 연간 보수교육 사업계획 자동화 구현
	return fmt.Errorf("연간 보수교육 사업계획 자동화는 아직 구현되지 않았습니다")
}
