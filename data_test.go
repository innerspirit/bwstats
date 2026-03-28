package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindReplayFilesInDir(t *testing.T) {
	tempDir := t.TempDir()
	dateDir := filepath.Join(tempDir, "2026-01-01")
	if err := os.Mkdir(dateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// create some test files
	if _, err := os.Create(filepath.Join(dateDir, "game1.rep")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Create(filepath.Join(dateDir, "notarep.txt")); err != nil {
		t.Fatal(err)
	}

	repFiles, err := findReplayFilesInDir(tempDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repFiles) != 1 {
		t.Fatalf("expected 1 .rep file, got %d", len(repFiles))
	}
}

func TestLoadUserNicknameFromPath(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "CSettings.json")
	content := `{"Gateway History":[{"account":"player123"}]}`
	if err := os.WriteFile(settingsPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	user, err := loadUserNicknameFromPath(settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "player123" {
		t.Fatalf("expected player123, got %q", user)
	}
}
