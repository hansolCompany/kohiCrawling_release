package main

import (
	"flag"
	"fmt"
	"os"

	"kohiCrawling/internal/api"
	"kohiCrawling/internal/updater"
	"kohiCrawling/server"
)

func main() {
	showVersion := flag.Bool("version", false, "현재 버전 출력")
	skipUpdate := flag.Bool("skip-update", false, "시작 시 업데이트 확인 건너뛰기")
	flag.Parse()

	if *showVersion {
		fmt.Println(server.CurrentVersion())
		return
	}

	if err := updater.CheckForUpdates(*skipUpdate, updater.Config{
		Version:   server.CurrentVersion(),
		UpdateURL: server.UpdateURL,
		EnvVar:    "KOHI_CRAWLING_SERVER_UPDATE_URL",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "업데이트 확인: %v\n", err)
	}

	s := api.NewServer()
	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "서버 오류: %v\n", err)
		os.Exit(1)
	}
}
