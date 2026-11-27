package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

const (
	liveVideo    = `{"@context":"https://schema.org","@graph":[{"@type":"VideoObject","description":"Live now","publication":{"@type":"BroadcastEvent","isLiveBroadcast":true}}]}`
	offlineVideo = `{"@context":"https://schema.org","@graph":[{"@type":"VideoObject","description":"Offline title","publication":{"@type":"BroadcastEvent","isLiveBroadcast":false}}]}`
	webPage      = `{"@context":"https://schema.org","@graph":[{"@type":"WebPage","name":"Channel"}]}`
)

func pageWithJSONLD(scripts ...string) string {
	var b strings.Builder
	b.WriteString("<html><body>")
	for _, script := range scripts {
		fmt.Fprintf(&b, `<script type="application/ld+json">%s</script>`, script)
	}
	b.WriteString("</body></html>")
	return b.String()
}

func parseHTML(t *testing.T, html string) (bool, string) {
	t.Helper()
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(html))}
	live, title, err := parse(resp)
	if err != nil {
		t.Fatalf("parse() error: %v", err)
	}
	return live, title
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		wantLive  bool
		wantTitle string
	}{
		{
			name:      "live stream",
			html:      pageWithJSONLD(liveVideo),
			wantLive:  true,
			wantTitle: "Live now",
		},
		{
			name:      "offline stream",
			html:      pageWithJSONLD(offlineVideo),
			wantLive:  false,
			wantTitle: "Offline title",
		},
		{
			name:      "no JSON-LD scripts",
			html:      "<html><body><p>Nothing useful here</p></body></html>",
			wantLive:  false,
			wantTitle: "",
		},
		{
			name:      "VideoObject in a later script",
			html:      pageWithJSONLD(webPage, liveVideo),
			wantLive:  true,
			wantTitle: "Live now",
		},
		{
			name:      "malformed script before a valid one",
			html:      pageWithJSONLD("{not json", liveVideo),
			wantLive:  true,
			wantTitle: "Live now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			live, title := parseHTML(t, tt.html)
			if live != tt.wantLive {
				t.Errorf("live = %t, want %t", live, tt.wantLive)
			}
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
		})
	}
}
