package main

import (
	"flag"
	"fmt"
	"os"

	"kohiCrawling/internal/credentials"
	"kohiCrawling/internal/updater"
	"kohiCrawling/kohi"
)

func main() {
	showVersion := flag.Bool("version", false, "현재 버전 출력")
	skipUpdate := flag.Bool("skip-update", false, "시작 시 업데이트 확인 건너뛰기")
	flag.Parse()

	if *showVersion {
		fmt.Println(kohi.CurrentVersion())
		return
	}

	if err := updater.CheckForUpdates(*skipUpdate, updater.Config{
		Version:   kohi.CurrentVersion(),
		UpdateURL: kohi.UpdateURL,
		EnvVar:    "KOHI_UPDATE_URL",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "업데이트 확인: %v\n", err)
	}

	creds := credentials.Read()

	if err := kohi.Run(creds); err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		os.Exit(1)
	}
}
