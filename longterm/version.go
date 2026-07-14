package longterm

// Version은 빌드 시 -ldflags로 주입됩니다.
var Version = "dev"

// UpdateURL은 update.json 위치입니다. 빌드 시 -ldflags로 덮어쓸 수 있습니다.
var UpdateURL = "https://example.com/longtermCrawling/update.json"

func CurrentVersion() string {
	if Version == "" {
		return "dev"
	}
	return Version
}
