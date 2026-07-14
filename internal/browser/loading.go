package browser

import (
	"errors"
	"time"

	"github.com/playwright-community/playwright-go"
)

func WaitLoading(page playwright.Page) (string, error) {
	loopCount := 0
	time.Sleep(500 * time.Millisecond)
	for {
		loading := page.Locator(".cl-control.cl-container.loadMask")
		loadingCount, err := loading.Count()

		if err != nil {
			return "로딩 마스크 확인 중 오류 발생", err
		}

		if loadingCount == 0 {
			break
		}

		allInvisible := true

		for i := 0; i < loadingCount; i++ {
			isVisible, err := loading.Nth(i).IsVisible()
			if err != nil {
				return "로딩 마스크 가시성 확인 중 오류 발생", err
			}

			if isVisible {
				allInvisible = false
				break
			}
		}

		if allInvisible {
			break
		}

		time.Sleep(500 * time.Millisecond)

		loopCount++
		if loopCount > 10 {
			return "로딩 대기 중 오류 발생, 루프 카운트 초과", errors.New("로딩 대기 중 오류 발생, 루프 카운트 초과")
		}
	}

	return "", nil
}
