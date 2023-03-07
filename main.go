package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

  checkCmd := flag.NewFlagSet("check", flag.ExitOnError)

  addCmd := flag.NewFlagSet("add", flag.ExitOnError)

  delCmd := flag.NewFlagSet("delete", flag.ExitOnError)


  if len(os.Args) < 2 {
    // print help
    fmt.Println("Not enough arguments!")
    os.Exit(1)
  }

  switch os.Args[1] {
  case "check":
    HandleCheck(checkCmd)
  case "add":
    HandleAdd(addCmd)
  case "del":
    HandleDel(delCmd)
  default:
    fmt.Println("Wrong Command!")
    os.Exit(1)
  }
}

func HandleCheck(checkCmd *flag.FlagSet) {
  checkStreamers()
}

func HandleAdd(addCmd *flag.FlagSet) {

  if len(os.Args) < 3 {
    fmt.Println("The streamer name is needed!")
    os.Exit(1)
  }

  addStreamer(os.Args[2:])
}

func HandleDel(delCmd *flag.FlagSet) {

  if len(os.Args) < 3 {
    fmt.Println("The streamer name is needed!")
    os.Exit(1)
  }

  delStreamer(os.Args[2:])
}
