#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

TARGET="${1:-kohi}"
VERSION="${2:-dev}"
DIST="$ROOT/dist"

build_current() {
  local name="$1"
  local pkg="$2"
  local ld_version="$3"
  local ld_url="$4"
  local update_url="${UPDATE_URL:-https://example.com/${name}/update.json}"
  local ldflags="-X ${ld_version}=${VERSION} -X ${ld_url}=${update_url}"
  echo "빌드: ${name}"
  go build -ldflags "${ldflags}" -o "${name}" "${pkg}"
  echo "  완료: ${name}"
}

build_all_platforms() {
  local name="$1"
  local pkg="$2"
  local ld_version="$3"
  local ld_url="$4"
  local update_url="${UPDATE_URL:-https://example.com/${name}/update.json}"
  local ldflags="-X ${ld_version}=${VERSION} -X ${ld_url}=${update_url}"
  mkdir -p "$DIST"

  echo "빌드: ${name} (windows/amd64, darwin/amd64, darwin/arm64)"
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "${ldflags}" -o "${DIST}/${name}.exe" "${pkg}"
  GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "${ldflags}" -o "${DIST}/${name}-darwin-amd64" "${pkg}"
  GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "${ldflags}" -o "${DIST}/${name}-darwin-arm64" "${pkg}"
}

case "$TARGET" in
  kohi)
    if [[ "${ALL_PLATFORMS:-}" == "1" ]]; then
      build_all_platforms "kohiCrawling" "./cmd/kohi" "kohiCrawling/kohi.Version" "kohiCrawling/kohi.UpdateURL"
    else
      build_current "kohiCrawling" "./cmd/kohi" "kohiCrawling/kohi.Version" "kohiCrawling/kohi.UpdateURL"
    fi
    ;;
  longterm)
    if [[ "${ALL_PLATFORMS:-}" == "1" ]]; then
      build_all_platforms "longtermCrawling" "./cmd/longterm" "kohiCrawling/longterm.Version" "kohiCrawling/longterm.UpdateURL"
    else
      build_current "longtermCrawling" "./cmd/longterm" "kohiCrawling/longterm.Version" "kohiCrawling/longterm.UpdateURL"
    fi
    ;;
  server)
    if [[ "${ALL_PLATFORMS:-}" == "1" ]]; then
      mkdir -p "$DIST"
      GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o "${DIST}/kohiCrawlingServer.exe" ./cmd/server
      GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o "${DIST}/kohiCrawlingServer-darwin-amd64" ./cmd/server
      GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o "${DIST}/kohiCrawlingServer-darwin-arm64" ./cmd/server
    else
      echo "빌드: kohiCrawlingServer"
      go build -o kohiCrawlingServer ./cmd/server
    fi
    ;;
  all)
    ALL_PLATFORMS=1 "$0" kohi "$VERSION"
    ALL_PLATFORMS=1 "$0" longterm "$VERSION"
    ALL_PLATFORMS=1 "$0" server "$VERSION"
    ;;
  *)
    echo "usage: ./build.sh [kohi|longterm|server|all] [version]"
    exit 1
    ;;
esac

echo ""
echo "Mac에서 실행 전 Playwright 설치: 프로그램 첫 실행 시 자동 설치됩니다."
