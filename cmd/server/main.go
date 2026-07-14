package main

import (
	"fmt"
	"os"

	"kohiCrawling/internal/api"
)

func main() {
	server := api.NewServer()
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "서버 오류: %v\n", err)
		os.Exit(1)
	}
}
