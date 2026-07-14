package longterm

import (
	"fmt"
	"strconv"

	"github.com/playwright-community/playwright-go"
)

const (
	scheduleMenuItemSelector = `[id="mainframe.VFrameSet.HFrameSet.frameLeft.form.grd_leftMenu.body.gridrow_7.cell_7_0.celltreeitem.treeitemtext"]`

	scheduleFormPrefix = "mainframe.VFrameSet.HFrameSet.VFrameSetSub.framesetWork.winNPA02310300.form._div_bizFrameMain.form"
	scheduleGridPrefix = scheduleFormPrefix + ".grd_jbEduAdminEduDyprList.body.gridrow_"

	instructorModalID            = "mainframe.VFrameSet.HFrameSet.VFrameSetSub.framesetWork.winNPA02310300.nprf356p01"
	instructorModalFormPrefix    = instructorModalID + ".form._div_bizFrameMain.form"
	instructorModalGridRowPrefix = instructorModalFormPrefix + ".grd_dyprList.body.gridrow_"
	scheduleSaveButtonID         = "mainframe.VFrameSet.HFrameSet.VFrameSetSub.framesetWork.winNPA02310300.form.div_functionButton.form.btn_save:icontext"
)

func locatorByID(page playwright.Page, id string) playwright.Locator {
	return page.Locator(`[id="` + id + `"]`)
}

func scheduleGridCellID(rowIndex, colIndex int, suffix string) string {
	return fmt.Sprintf("%s.grd_jbEduAdminEduDyprList.body.gridrow_%d.cell_%d_%d.%s",
		scheduleFormPrefix, rowIndex, rowIndex, colIndex, suffix)
}

func scheduleGridCellDivID(rowIndex, colIndex int) string {
	return fmt.Sprintf("%s.grd_jbEduAdminEduDyprList.body.gridrow_%d.cell_%d_%d",
		scheduleFormPrefix, rowIndex, rowIndex, colIndex)
}

func scheduleMenuItem(page playwright.Page) playwright.Locator {
	return page.Locator(scheduleMenuItemSelector)
}

func scheduleSearchBeginInput(page playwright.Page) playwright.Locator {
	return locatorByID(page, scheduleFormPrefix+".cal_eduDt.form.cal_begin.calendaredit:input")
}

func scheduleSearchEndInput(page playwright.Page) playwright.Locator {
	return locatorByID(page, scheduleFormPrefix+".cal_eduDt.form.cal_end.calendaredit:input")
}

func scheduleSelectButton(page playwright.Page) playwright.Locator {
	return locatorByID(page, scheduleFormPrefix+".btn_select:icontext")
}

func scheduleRowAddButton(page playwright.Page) playwright.Locator {
	return locatorByID(page, scheduleFormPrefix+".btn_rowAdd:icontext")
}

func clickScheduleMenu(page playwright.Page) error {
	menu := scheduleMenuItem(page)
	if err := ClickWithLoading(page, menu, 30000); err != nil {
		return fmt.Errorf("보수교육 일정관리 메뉴 클릭 실패: %w", err)
	}

	fmt.Println("보수교육 일정관리 메뉴 클릭 완료")
	return nil
}

func findEmptyScheduleGridRow(page playwright.Page) (int, playwright.Locator, error) {
	result, err := page.Evaluate(`(gridRowPrefix) => {
		const isFilledDate = (text) => /^\d{4}-\d{2}-\d{2}$/.test(text.replace(/\s/g, ''));
		const isEmptyPlaceholder = (text) => {
			const value = (text || '').trim();
			if (!value) {
				return true;
			}
			if (isFilledDate(value)) {
				return false;
			}
			return /^[\s-]+$/.test(value);
		};

		for (const row of document.querySelectorAll('[id^="' + gridRowPrefix + '"]')) {
			const suffix = row.id.slice(gridRowPrefix.length);
			const match = suffix.match(/^(\d+)$/);
			if (!match) {
				continue;
			}

			const n = match[1];
			const cellId = gridRowPrefix + n + '.cell_' + n + '_2.cellcalendar.calendaredit:input';
			const cell = document.getElementById(cellId);
			if (!cell) {
				continue;
			}

			const text = cell.innerText || cell.textContent || '';
			if (isEmptyPlaceholder(text)) {
				return { rowIndex: parseInt(n, 10), cellId };
			}
		}

		return null;
	}`, scheduleGridPrefix)
	if err != nil {
		return 0, nil, fmt.Errorf("빈 교육일자 행 검색 실패: %w", err)
	}

	if result == nil {
		return 0, nil, fmt.Errorf("빈 교육일자 입력란을 찾을 수 없음")
	}

	rowData, ok := result.(map[string]interface{})
	if !ok {
		return 0, nil, fmt.Errorf("빈 교육일자 행 결과 파싱 실패")
	}

	rowIndex, ok := evalInt(rowData["rowIndex"])
	if !ok {
		return 0, nil, fmt.Errorf("행 index 파싱 실패")
	}

	cellID, ok := rowData["cellId"].(string)
	if !ok || cellID == "" {
		return 0, nil, fmt.Errorf("교육일자 입력란 id 파싱 실패")
	}

	return rowIndex, locatorByID(page, cellID), nil
}

func evalInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

func findGridComboListItem(page playwright.Page, rowIndex, colIndex int, optionText string, matchContains bool) (playwright.Locator, error) {
	itemPrefix := scheduleGridPrefix + strconv.Itoa(rowIndex) + ".cell_" + strconv.Itoa(rowIndex) +
		"_" + strconv.Itoa(colIndex) + ".cellcombo.combolist.item_"

	result, err := page.Evaluate(`({ itemPrefix, optionText, matchContains }) => {
		for (const el of document.querySelectorAll('[id^="' + itemPrefix + '"]')) {
			if (!el.id.endsWith(':text')) {
				continue;
			}
			const text = (el.innerText || el.textContent || '').trim();
			const matched = matchContains ? text.includes(optionText) : text === optionText;
			if (matched) {
				return el.id;
			}
		}
		return null;
	}`, map[string]interface{}{
		"itemPrefix":    itemPrefix,
		"optionText":    optionText,
		"matchContains": matchContains,
	})
	if err != nil {
		return nil, fmt.Errorf("콤보 항목 검색 실패: %w", err)
	}

	itemID, ok := result.(string)
	if !ok || itemID == "" {
		return nil, fmt.Errorf("콤보 항목 '%s'을(를) 찾을 수 없음", optionText)
	}

	return locatorByID(page, itemID), nil
}

func clickGridComboOption(page playwright.Page, rowIndex, colIndex int, optionText string, matchContains bool) error {
	cell := locatorByID(page, scheduleGridCellDivID(rowIndex, colIndex))
	if err := ClickWithLoadingFast(page, cell, 30000); err != nil {
		return fmt.Errorf("셀 클릭 실패 (row %d, col %d): %w", rowIndex, colIndex, err)
	}

	dropButton := locatorByID(page, scheduleGridCellID(rowIndex, colIndex, "cellcombo.dropbutton:icontext"))
	if err := ClickWithLoadingFast(page, dropButton, 30000); err != nil {
		return fmt.Errorf("콤보 dropbutton 클릭 실패 (row %d, col %d): %w", rowIndex, colIndex, err)
	}

	item, err := findGridComboListItem(page, rowIndex, colIndex, optionText, matchContains)
	if err != nil {
		return err
	}

	if err := ClickWithLoadingFast(page, item, 30000); err != nil {
		return fmt.Errorf("콤보 항목 '%s' 클릭 실패: %w", optionText, err)
	}

	fmt.Printf("  콤보 선택 완료 (row %d, col %d): %s\n", rowIndex, colIndex, optionText)
	return nil
}

func findInstructorModalItem(page playwright.Page, instructorName string) (playwright.Locator, error) {
	result, err := page.Evaluate(`({ gridRowPrefix, instructorName }) => {
		for (let m = 0; ; m++) {
			const rowId = gridRowPrefix + m;
			const row = document.getElementById(rowId);
			if (!row) {
				break;
			}

			const cellId = gridRowPrefix + m + '.cell_' + m + '_1:text';
			const cell = document.getElementById(cellId);
			if (!cell) {
				continue;
			}

			const text = (cell.innerText || cell.textContent || '').trim();
			if (text === instructorName) {
				return cellId;
			}
		}
		return null;
	}`, map[string]interface{}{
		"gridRowPrefix":  instructorModalGridRowPrefix,
		"instructorName": instructorName,
	})
	if err != nil {
		return nil, fmt.Errorf("강사 목록 검색 실패: %w", err)
	}

	itemID, ok := result.(string)
	if !ok || itemID == "" {
		return nil, fmt.Errorf("강사 '%s'을(를) 찾을 수 없음", instructorName)
	}

	return locatorByID(page, itemID), nil
}

func selectInstructor(page playwright.Page, rowIndex int, instructorName string) error {
	instructorCell := locatorByID(page, scheduleGridCellDivID(rowIndex, 7))
	if err := ClickWithLoadingFast(page, instructorCell, 30000); err != nil {
		return fmt.Errorf("강사 선택 셀 클릭 실패: %w", err)
	}

	modal := locatorByID(page, instructorModalID)
	if err := modal.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		return fmt.Errorf("강사 선택 모달 표시 대기 실패: %w", err)
	}

	instructorItem, err := findInstructorModalItem(page, instructorName)
	if err != nil {
		return err
	}

	if err := ClickWithLoadingFast(page, instructorItem, 30000); err != nil {
		return fmt.Errorf("강사 '%s' 클릭 실패: %w", instructorName, err)
	}

	choiceButton := locatorByID(page, instructorModalFormPrefix+".btn_choice:icontext")
	if err := ClickWithLoadingFast(page, choiceButton, 30000); err != nil {
		return fmt.Errorf("강사 선택 확인 버튼 클릭 실패: %w", err)
	}

	fmt.Printf("  강사 선택 완료: %s\n", instructorName)
	return nil
}

func saveSchedule(page playwright.Page) error {
	saveButton := locatorByID(page, scheduleSaveButtonID)
	if err := ClickWithLoadingFast(page, saveButton, 30000); err != nil {
		return fmt.Errorf("저장 버튼 클릭 실패: %w", err)
	}

	fmt.Println("  일정 저장 완료")
	return nil
}

func fillScheduleGridRow(page playwright.Page, rowIndex int, req ScheduleRequest) error {
	if err := clickGridComboOption(page, rowIndex, 3, req.StandardArea, false); err != nil {
		return fmt.Errorf("표준영역 선택 실패: %w", err)
	}

	if err := clickGridComboOption(page, rowIndex, 4, req.CourseName, true); err != nil {
		return fmt.Errorf("과목명 선택 실패: %w", err)
	}

	startTimeInput := locatorByID(page, scheduleGridCellID(rowIndex, 5, "cellmaskedit:input"))
	if err := TypeSlowlyWithLoadingFast(page, startTimeInput, req.CourseStartTime, 30000); err != nil {
		return fmt.Errorf("과목시작시간 입력 실패: %w", err)
	}

	endTimeInput := locatorByID(page, scheduleGridCellID(rowIndex, 6, "cellmaskedit:input"))
	if err := TypeSlowlyWithLoadingFast(page, endTimeInput, req.CourseEndTime, 30000); err != nil {
		return fmt.Errorf("과목종료시간 입력 실패: %w", err)
	}

	if err := selectInstructor(page, rowIndex, req.InstructorName); err != nil {
		return fmt.Errorf("강사 선택 실패: %w", err)
	}

	capacityCell := locatorByID(page, scheduleGridCellID(rowIndex, 8, "cellmaskedit"))
	if err := ClickWithLoadingFast(page, capacityCell, 30000); err != nil {
		return fmt.Errorf("정원 셀 클릭 실패: %w", err)
	}

	capacityInput := locatorByID(page, scheduleGridCellID(rowIndex, 8, "cellmaskedit:input"))
	if err := TypeSlowlyWithLoadingFast(page, capacityInput, strconv.Itoa(req.Capacity), 30000); err != nil {
		return fmt.Errorf("정원 입력 실패: %w", err)
	}

	noteCell := locatorByID(page, scheduleGridCellDivID(rowIndex, 9))
	if err := ClickWithLoadingFast(page, noteCell, 30000); err != nil {
		return fmt.Errorf("비고 셀 클릭 실패: %w", err)
	}

	if req.Note != "" {
		noteInput := locatorByID(page, scheduleGridCellID(rowIndex, 9, "celledit:input"))
		if err := TypeSlowlyWithLoadingFast(page, noteInput, req.Note, 30000); err != nil {
			return fmt.Errorf("비고 입력 실패: %w", err)
		}
		fmt.Printf("  비고 입력 완료: %s\n", req.Note)
	}

	if err := saveSchedule(page); err != nil {
		return err
	}

	return nil
}

func prepareScheduleRow(page playwright.Page, req ScheduleRequest) error {
	beginDate, err := educationDateMinusOneMonth(req.EducationDate)
	if err != nil {
		return err
	}
	endDate, err := educationDatePlusOneMonth(req.EducationDate)
	if err != nil {
		return err
	}

	fmt.Printf("  검색 시작일: %s, 종료일: %s\n", beginDate, endDate)

	if err := TypeSlowlyWithLoadingFast(page, scheduleSearchBeginInput(page), beginDate, 30000); err != nil {
		return fmt.Errorf("교육일자 시작 입력 실패: %w", err)
	}

	if err := TypeSlowlyWithLoadingFast(page, scheduleSearchEndInput(page), endDate, 30000); err != nil {
		return fmt.Errorf("교육일자 종료 입력 실패: %w", err)
	}

	if err := ClickWithLoadingFast(page, scheduleSelectButton(page), 30000); err != nil {
		return fmt.Errorf("조회 버튼 클릭 실패: %w", err)
	}

	if err := ClickWithLoadingFast(page, scheduleRowAddButton(page), 30000); err != nil {
		return fmt.Errorf("행 추가 버튼 클릭 실패: %w", err)
	}

	rowIndex, dateInput, err := findEmptyScheduleGridRow(page)
	if err != nil {
		return err
	}

	if err := TypeSlowlyWithLoadingFast(page, dateInput, req.EducationDate, 30000); err != nil {
		return fmt.Errorf("그리드 교육일자 입력 실패: %w", err)
	}
	fmt.Printf("  교육일자 입력 완료: %s\n", req.EducationDate)

	if err := fillScheduleGridRow(page, rowIndex, req); err != nil {
		return err
	}

	return nil
}
