package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConstructConfigPath(t *testing.T) {
	home := withTempHome(t)

	want := filepath.Join(home, ".config", "ttvchecker")
	if got := constructConfigPath(); got != want {
		t.Fatalf("constructConfigPath() = %q, want %q", got, want)
	}
}

func TestCreateStreamerlist(t *testing.T) {
	withTempHome(t)
	path := filepath.Join(constructConfigPath(), "streamerlist.txt")

	if checkFileExist(path) {
		t.Fatal("streamerlist.txt exists before createStreamerlist")
	}

	createStreamerlist()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat created list: %v", err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("created list is not a regular file: %v", info.Mode())
	}

	// A later call must be idempotent and must not truncate existing entries.
	if err := os.WriteFile(path, []byte("foo\n"), 0o600); err != nil {
		t.Fatalf("write list: %v", err)
	}
	createStreamerlist()
	if got := readStreamerList(t); !reflect.DeepEqual(got, []string{"foo"}) {
		t.Fatalf("list after second createStreamerlist = %v, want [foo]", got)
	}
}

func TestOpenStreamerlistAndInitStreamerList(t *testing.T) {
	withTempHome(t)
	writeStreamerList(t, "foo\nbar\n")

	f := openStreamerlist()
	defer f.Close()

	wantPath := filepath.Join(constructConfigPath(), "streamerlist.txt")
	if f.Name() != wantPath {
		t.Fatalf("openStreamerlist() opened %q, want %q", f.Name(), wantPath)
	}

	want := []string{"foo", "bar"}
	if got := initStreamerList(); !reflect.DeepEqual(got, want) {
		t.Fatalf("initStreamerList() = %v, want %v", got, want)
	}
}

func TestCheckFileExist(t *testing.T) {
	dir := t.TempDir()
	if !checkFileExist(dir) {
		t.Fatal("checkFileExist returned false for an existing directory")
	}
	if checkFileExist(filepath.Join(dir, "missing")) {
		t.Fatal("checkFileExist returned true for a missing path")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		list  []string
		value string
		want  bool
	}{
		{name: "empty list", list: nil, value: "foo", want: false},
		{name: "present", list: []string{"foo", "bar"}, value: "bar", want: true},
		{name: "absent", list: []string{"foo", "bar"}, value: "baz", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.list, tt.value); got != tt.want {
				t.Fatalf("contains(%v, %q) = %t, want %t", tt.list, tt.value, got, tt.want)
			}
		})
	}
}

func TestFindMin(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{a: 1, b: 2, want: 1},
		{a: 2, b: 1, want: 1},
		{a: 3, b: 3, want: 3},
	}

	for _, tt := range tests {
		if got := findMin(tt.a, tt.b); got != tt.want {
			t.Fatalf("findMin(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestGetResponse(t *testing.T) {
	t.Run("success returns body and sets user agent", func(t *testing.T) {
		userAgents := make(chan string, 1)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userAgents <- r.UserAgent()
			_, _ = io.WriteString(w, "hello")
		}))
		defer srv.Close()

		resp, err := getResponse(srv.URL)
		if err != nil {
			t.Fatalf("getResponse() error: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read response body: %v", err)
		}
		if string(body) != "hello" {
			t.Fatalf("body = %q, want %q", body, "hello")
		}
		if got := <-userAgents; got != userAgent {
			t.Fatalf("User-Agent = %q, want %q", got, userAgent)
		}
	})

	t.Run("non-OK status is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		resp, err := getResponse(srv.URL)
		if err == nil {
			t.Fatal("getResponse() unexpectedly succeeded for a 404")
		}
		if resp != nil {
			t.Fatalf("getResponse() returned a response with an error: %v", resp)
		}
		if !strings.Contains(err.Error(), "404") {
			t.Fatalf("error = %q, want it to mention the status code", err)
		}
	})

	t.Run("invalid URL is an error", func(t *testing.T) {
		resp, err := getResponse("://not-a-url")
		if err == nil {
			t.Fatal("getResponse() unexpectedly succeeded for an invalid URL")
		}
		if resp != nil {
			t.Fatalf("getResponse() returned a response with an error: %v", resp)
		}
	})
}
