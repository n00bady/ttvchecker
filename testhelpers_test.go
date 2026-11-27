package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempHome points config lookup at an isolated directory for the duration
// of a test. os.UserHomeDir reads HOME on Unix and USERPROFILE on Windows.
func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	// createStreamerlist uses os.Mkdir (not MkdirAll), so the parent must exist.
	if err := os.Mkdir(filepath.Join(home, ".config"), 0o700); err != nil {
		t.Fatalf("create .config directory: %v", err)
	}
	return home
}

func writeStreamerList(t *testing.T, content string) {
	t.Helper()
	createStreamerlist()
	path := filepath.Join(constructConfigPath(), "streamerlist.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write streamer list: %v", err)
	}
}

func readStreamerList(t *testing.T) []string {
	t.Helper()
	path := filepath.Join(constructConfigPath(), "streamerlist.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read streamer list: %v", err)
	}
	text := strings.TrimSuffix(string(data), "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

// captureStdout collects output written to os.Stdout while fn runs. Tests that
// use it must not run in parallel because os.Stdout is global.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	original := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = original }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close pipe reader: %v", err)
	}
	return string(out)
}
