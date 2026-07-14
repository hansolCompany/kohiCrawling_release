package longterm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFileWithNameKeyword(t *testing.T) {
	dir := t.TempDir()

	oldPath := filepath.Join(dir, "보수교육 수강내역 등록 양식.xlsx")
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old file: %v", err)
	}

	newPath := filepath.Join(dir, "2026_보수교육 수강내역 등록.xlsx")
	if err := os.WriteFile(newPath, []byte("new"), 0o644); err != nil {
		t.Fatalf("write new file: %v", err)
	}

	found, err := findFileWithNameKeyword(enrollUploadFileNameKeyword, dir)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if found != newPath {
		t.Fatalf("found = %q, want %q", found, newPath)
	}
}

func TestUserDesktopDir(t *testing.T) {
	dir, err := userDesktopDir()
	if err != nil {
		t.Fatalf("userDesktopDir failed: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat desktop dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("desktop path is not a directory: %s", dir)
	}
}
