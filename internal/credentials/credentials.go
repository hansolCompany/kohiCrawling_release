package credentials

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Credentials struct {
	UserID   string
	Password string
}

func Read() Credentials {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("아이디: ")
	userID, _ := reader.ReadString('\n')

	fmt.Print("비밀번호: ")
	password, _ := reader.ReadString('\n')

	return Credentials{
		UserID:   strings.TrimSpace(userID),
		Password: strings.TrimSpace(password),
	}
}
