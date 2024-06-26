package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/alexeyco/simpletable"
)

// colors
var (
	reset     = "\033[0m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	// blue      = "\033[34m"
	// purple    = "\033[35m"
	// cyan      = "\033[36m"
	// gray      = "\033[37m"
	// dark_gray = "\x1b[38;5;236m"
	// white     = "\033[97m"
)

// gives color to output and prints the results
// take a slice of stream prints to output and
// returns an error or nil if no errors
func pPrint(s []stream) error {
	if len(s) == 0 {
		return errors.New("input slice is empty")
	}

	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "#"},
			{Align: simpletable.AlignLeft, Text: "Streamer"},
			{Align: simpletable.AlignLeft, Text: "Live?"},
			{Align: simpletable.AlignLeft, Text: "URL"},
		},
	}

	for i, n := range s {
		switch n.live {
		case true:
			r := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i)},
				{Text: n.name},
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%sLive!%s", green, reset)},
				{Text: n.url},
			}
			table.Body.Cells = append(table.Body.Cells, r)
		case false:
			r := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i)},
				{Text: n.name},
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%sOffline%s", red, reset)},
				{Text: n.url},
			}
			table.Body.Cells = append(table.Body.Cells, r)
		default:
			return errors.New("unexpected error on reading the bool from stream struct")
		}
	}

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())

	return nil
}

func csvPrint(s []stream) error {
	stream_string := make([][]string, len(s))

	for l, stream := range s {
		values := reflect.ValueOf(stream)
		for i := 0; i < values.NumField(); i++ {
			switch values.Field(i).Kind() {
			case reflect.Bool:
				// i could not find another way to print true/false instead of <bool value> except this
				stream_string[l] = append(stream_string[l], fmt.Sprintf("%v", values.Field(i).Bool()))
			default:
				stream_string[l] = append(stream_string[l], values.Field(i).String())
			}
		}
	}

	output := csv.NewWriter(os.Stdout)
	output.WriteAll(stream_string)
	output.Flush()

	return nil
}
