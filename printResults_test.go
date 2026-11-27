package main

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

func TestPPrint(t *testing.T) {
	streams := []stream{
		{name: "foo", live: true, title: "Title A", url: url + "foo"},
		{name: "bar", live: false, title: "Title B", url: url + "bar"},
	}

	out := captureStdout(t, func() {
		if err := pPrint(streams); err != nil {
			t.Errorf("pPrint() error: %v", err)
		}
	})

	for _, want := range []string{"foo", "bar", "Live!", "Offline", url + "foo"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output does not contain %q:\n%s", want, out)
		}
	}
}

func TestPPrintEmptySlice(t *testing.T) {
	out := captureStdout(t, func() {
		if err := pPrint(nil); err == nil {
			t.Error("pPrint(nil) unexpectedly succeeded")
		}
	})
	if out != "" {
		t.Errorf("pPrint(nil) wrote output: %q", out)
	}
}

func TestCSVPrint(t *testing.T) {
	streams := []stream{
		{name: "foo", live: true, title: "Title, with comma", url: url + "foo"},
		{name: "bar", live: false, title: "Title B", url: url + "bar"},
	}

	out := captureStdout(t, func() {
		if err := csvPrint(streams); err != nil {
			t.Errorf("csvPrint() error: %v", err)
		}
	})

	records, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV output: %v\noutput: %q", err, out)
	}
	want := [][]string{
		{"foo", "true", "Title, with comma", url + "foo"},
		{"bar", "false", "Title B", url + "bar"},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("CSV records = %v, want %v", records, want)
	}
}

func TestCSVPrintEmptySlice(t *testing.T) {
	out := captureStdout(t, func() {
		if err := csvPrint(nil); err != nil {
			t.Errorf("csvPrint(nil) error: %v", err)
		}
	})
	if out != "" {
		t.Errorf("csvPrint(nil) wrote output: %q", out)
	}
}
