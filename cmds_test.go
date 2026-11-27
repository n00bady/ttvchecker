package main

import (
	"reflect"
	"testing"
)

func TestAddStreamer(t *testing.T) {
	withTempHome(t)
	writeStreamerList(t, "existing\n")

	addStreamer([]string{"foo", "bar"})

	want := []string{"existing", "foo", "bar"}
	if got := readStreamerList(t); !reflect.DeepEqual(got, want) {
		t.Fatalf("list after add = %v, want %v", got, want)
	}
}

func TestDelStreamer(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		delete  []string
		want    []string
	}{
		{
			name:    "deletes one name",
			initial: "foo\nbar\n",
			delete:  []string{"foo"},
			want:    []string{"bar"},
		},
		{
			name:    "deletes several names",
			initial: "foo\nbar\nbaz\n",
			delete:  []string{"foo", "baz"},
			want:    []string{"bar"},
		},
		{
			name:    "missing name leaves list unchanged",
			initial: "foo\nbar\n",
			delete:  []string{"baz"},
			want:    []string{"foo", "bar"},
		},
		{
			name:    "deleting every name leaves an empty list",
			initial: "foo\n",
			delete:  []string{"foo"},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTempHome(t)
			writeStreamerList(t, tt.initial)

			delStreamer(tt.delete)

			if got := readStreamerList(t); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("list after delete = %v, want %v", got, tt.want)
			}
		})
	}
}
