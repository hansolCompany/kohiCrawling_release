package browser

import (
	"fmt"
	"sync"

	"github.com/playwright-community/playwright-go"
)

var dialogTracker struct {
	sync.Mutex
	lastMessage string
}

// SetupDialogHandler는 브라우저 전체에서 뜨는 alert/confirm/prompt를 자동 처리한다.
func SetupDialogHandler(ctx playwright.BrowserContext) {
	ctx.OnDialog(func(dialog playwright.Dialog) {
		msg := dialog.Message()
		dialogTracker.Lock()
		dialogTracker.lastMessage = msg
		dialogTracker.Unlock()

		fmt.Printf("알림: %s\n", msg)
		_ = dialog.Accept()
	})
}

func LastDialogMessage() string {
	dialogTracker.Lock()
	defer dialogTracker.Unlock()
	return dialogTracker.lastMessage
}

func ClearDialogMessage() {
	dialogTracker.Lock()
	dialogTracker.lastMessage = ""
	dialogTracker.Unlock()
}
