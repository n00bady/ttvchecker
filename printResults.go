package main

import (
	"errors"
	"fmt"

	"github.com/alexeyco/simpletable"
)

// colors
var reset  = "\033[0m"
var red    = "\033[31m"
var green  = "\033[32m"
var yellow = "\033[33m"
var blue   = "\033[34m"
var purple = "\033[35m"
var cyan   = "\033[36m"
var gray   = "\033[37m"
var white  = "\033[97m"

// gives color to output and prints the results
// take a slice of stream prints to output and
// returns an error or nil if no errors
func pPrint(s []stream) error {

  if len(s) == 0 {
    return errors.New("Input slice is empty!")
  }

  table := simpletable.New()

  table.Header = &simpletable.Header{
    Cells: []*simpletable.Cell{
      {Align: simpletable.AlignLeft, Text: "#"},
      {Align: simpletable.AlignLeft, Text: "Streamer"},
      {Align: simpletable.AlignLeft, Text: "Live?"},
    },
  }

  for i, n := range s {
    switch n.live {
    case true:
      r := []*simpletable.Cell{
        {Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i)},
        {Text: n.name},
        {Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s%t%s",green, n.live, reset)},
      }
    table.Body.Cells = append(table.Body.Cells, r)
    case false:
      r := []*simpletable.Cell{
        {Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", i)},
        {Text: n.name},
        {Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s%t%s",red, n.live, reset)},
      }
    table.Body.Cells = append(table.Body.Cells, r)
    default:
      return errors.New("Unexpected error on reading the bool from stream struct.")
    }
  }

  table.SetStyle(simpletable.StyleCompactLite)
  fmt.Println(table.String())

  return nil
}
