package main

// Regression tests for known bugs. These are expected to FAIL until the
// corresponding production fixes are made. They are intentionally kept out of
// the committed test suite.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// parse currently lets a later VideoObject block overwrite an earlier one.
func TestParseUsesFirstVideoObject(t *testing.T) {
	live := `{"@graph":[{"@type":"VideoObject","description":"Live now","publication":{"isLiveBroadcast":true}}]}`
	cached := `{"@graph":[{"@type":"VideoObject","description":"Cached title","publication":{"isLiveBroadcast":false}}]}`

	gotLive, gotTitle := parseHTML(t, pageWithJSONLD(live, cached))
	if !gotLive {
		t.Error("parse reported offline when the first VideoObject was live")
	}
	if gotTitle != "Live now" {
		t.Errorf("title = %q, want the first VideoObject's title %q", gotTitle, "Live now")
	}
}

// createStreamerlist uses os.Mkdir instead of os.MkdirAll, so it fails when
// ~/.config does not already exist. Run in a subprocess because the production
// code calls os.Exit on failure.
func TestCreateStreamerlistWithoutConfigParent(t *testing.T) {
	const flag = "TTVCHECKER_TEST_CREATE_CONFIG"
	if os.Getenv(flag) == "1" {
		home := os.Getenv("TTVCHECKER_TEST_HOME")
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		createStreamerlist()
		path := filepath.Join(home, ".config", "ttvchecker", "streamerlist.txt")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("list not created: %v", err)
		}
		return
	}

	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCreateStreamerlistWithoutConfigParent$")
	cmd.Env = append(os.Environ(), flag+"=1", "TTVCHECKER_TEST_HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("createStreamerlist failed when ~/.config was missing (%v):\n%s", err, out)
	}
}

// addStreamer writes raw arguments: whitespace, URLs, duplicates and blanks.
func TestAddStreamerNormalizesNames(t *testing.T) {
	withTempHome(t)
	writeStreamerList(t, "")

	addStreamer([]string{"  Foo  ", "https://www.twitch.tv/foo", "foo", ""})

	want := []string{"foo"}
	if got := readStreamerList(t); !reflect.DeepEqual(got, want) {
		t.Fatalf("list = %v, want normalized and deduplicated %v", got, want)
	}
}

// delStreamer compares logins case-sensitively.
func TestDelStreamerCaseInsensitive(t *testing.T) {
	withTempHome(t)
	writeStreamerList(t, "foo\nbar\n")

	delStreamer([]string{"FOO"})

	want := []string{"bar"}
	if got := readStreamerList(t); !reflect.DeepEqual(got, want) {
		t.Fatalf("list = %v, want %v", got, want)
	}
}

// getResponse does not close the body when the status code is not 200.
type closeTrackingBody struct {
	io.ReadCloser
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return b.ReadCloser.Close()
}

type trackingTransport struct {
	base http.RoundTripper
	body *closeTrackingBody
}

func (rt *trackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.base.RoundTrip(req)
	if err != nil || resp.StatusCode == http.StatusOK {
		return resp, err
	}
	rt.body = &closeTrackingBody{ReadCloser: resp.Body}
	resp.Body = rt.body
	return resp, err
}

func TestGetResponseClosesBodyOnNonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	original := http.DefaultTransport
	tracker := &trackingTransport{base: original}
	http.DefaultTransport = tracker
	defer func() { http.DefaultTransport = original }()

	resp, err := getResponse(srv.URL)
	if err == nil {
		t.Fatal("getResponse unexpectedly succeeded")
	}
	defer func() {
		if tracker.body != nil {
			tracker.body.Close()
		}
	}()
	if resp != nil {
		t.Fatal("getResponse returned a response alongside the error")
	}
	if tracker.body == nil {
		t.Fatal("test transport did not see the response")
	}
	if !tracker.body.closed {
		t.Error("getResponse left the response body open after a 404")
	}
}

// Update dereferences SelectedRow() without checking whether any rows exist.
func TestTUIEnterWithNoRows(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update panicked on enter with an empty table: %v", r)
		}
	}()

	m := model{
		table:        table.New(table.WithColumns([]table.Column{{Title: "#", Width: 3}})),
		streamerlist: []string{"foo"},
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
}

// Init assumes at least one streamer is in the list.
func TestTUIInitWithEmptyStreamerList(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init panicked with an empty streamer list: %v", r)
		}
	}()

	m := model{}
	m.Init()
}
