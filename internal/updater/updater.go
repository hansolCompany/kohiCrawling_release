package updater

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/minio/selfupdate"
)

type Config struct {
	Version   string
	UpdateURL string
	EnvVar    string
}

type platformAsset struct {
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
}

type updateManifest struct {
	Version     string                    `json:"version"`
	DownloadURL string                    `json:"download_url"`
	SHA256      string                    `json:"sha256"`
	Assets      map[string]platformAsset `json:"assets,omitempty"`
}

func (m *updateManifest) resolveDownload() (downloadURL, sha256 string, err error) {
	key := platformKey()
	if m.Assets != nil {
		if asset, ok := m.Assets[key]; ok && asset.DownloadURL != "" {
			return asset.DownloadURL, asset.SHA256, nil
		}
	}

	if m.DownloadURL != "" {
		return m.DownloadURL, m.SHA256, nil
	}

	return "", "", fmt.Errorf("update.json에 %s용 download_url이 없습니다", key)
}

// CheckForUpdates는 원격 manifest와 현재 버전을 비교해 새 버전이 있으면 다운로드·적용 후 재시작합니다.
func CheckForUpdates(skipUpdate bool, cfg Config) error {
	if skipUpdate {
		fmt.Println("업데이트 확인을 건너뜁니다.")
		return nil
	}

	updateURL := strings.TrimSpace(os.Getenv(cfg.EnvVar))
	if updateURL == "" {
		updateURL = strings.TrimSpace(cfg.UpdateURL)
	}
	if updateURL == "" || strings.Contains(updateURL, "example.com") {
		fmt.Printf("현재 버전: %s (업데이트 URL 미설정)\n", cfg.Version)
		return nil
	}

	manifest, err := fetchUpdateManifest(updateURL)
	if err != nil {
		return fmt.Errorf("업데이트 정보 조회 실패: %w", err)
	}

	current, err := version.NewVersion(cfg.Version)
	if err != nil {
		return fmt.Errorf("현재 버전 파싱 실패: %w", err)
	}

	latest, err := version.NewVersion(manifest.Version)
	if err != nil {
		return fmt.Errorf("원격 버전 파싱 실패: %w", err)
	}

	fmt.Printf("현재 버전: %s / 최신 버전: %s\n", current, latest)

	if !latest.GreaterThan(current) {
		fmt.Println("최신 버전을 사용 중입니다.")
		return nil
	}

	fmt.Printf("새 버전(%s)을 다운로드합니다...\n", latest)

	if err := applyUpdate(manifest); err != nil {
		return fmt.Errorf("업데이트 적용 실패: %w", err)
	}

	fmt.Println("업데이트 완료. 프로그램을 재시작합니다.")
	return restartExecutable()
}

func fetchUpdateManifest(updateURL string) (*updateManifest, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(updateURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d (%s)", resp.StatusCode, updateURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	body = bytes.TrimPrefix(body, []byte{0xEF, 0xBB, 0xBF})

	manifest := &updateManifest{}
	if err := json.Unmarshal(body, manifest); err != nil {
		return nil, err
	}

	if manifest.Version == "" {
		return nil, fmt.Errorf("update.json에 version이 없습니다")
	}
	if _, _, err := manifest.resolveDownload(); err != nil {
		return nil, err
	}

	return manifest, nil
}

func platformKey() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

func applyUpdate(manifest *updateManifest) error {
	downloadURL, sha256Hex, err := manifest.resolveDownload()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("다운로드 HTTP %d", resp.StatusCode)
	}

	opts := selfupdate.Options{}
	if sha256Hex != "" {
		sum, err := hex.DecodeString(strings.TrimSpace(sha256Hex))
		if err != nil {
			return fmt.Errorf("sha256 파싱 실패: %w", err)
		}
		opts.Checksum = sum
	}

	return selfupdate.Apply(resp.Body, opts)
}

func restartExecutable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, filterRestartArgs(os.Args[1:])...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func filterRestartArgs(args []string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--skip-update" || arg == "-skip-update" {
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}
