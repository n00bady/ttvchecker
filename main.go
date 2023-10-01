package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Make my own usage message
	usgmsg := "Commands:\n" +
		"\t check \tCheck if the streams are online.\n" +
		"\t\t  -l \tShow only the streams that are live.\n" +
		"\t\t --csv \tGet the output in csv format.\n\n" +
		"\t add   \tAdd one or more streamers in the list.\n" +
		"\t del   \tDelete one or more streamers from the list.\n" +
		"\t tui   \tStart the TUI!\n"

	// Initialize subcommands and possibly their options in the future(?)
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	onlyLives := checkCmd.Bool("l", false, "Show only streams that are currently live.")
	formatCSV := checkCmd.Bool("csv", false, "Return the results in a comma seperated format.")

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)

	delCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	tuiCmd := flag.NewFlagSet("tui", flag.ExitOnError)

	// Customize flag.Usage() with our own message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [command] [args]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, usgmsg)
	}

	// parse global options (they do not exists... yet)
	flag.Parse()
	subcmd := flag.Args()
	if len(subcmd) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	switch subcmd[0] {
	case "check":
		checkCmd.Parse(os.Args[2:])
		HandleCheck(checkCmd, *onlyLives, *formatCSV)
	case "add":
		HandleAdd(addCmd, subcmd[1:])
	case "del":
		HandleDel(delCmd, subcmd[1:])
	case "tui":
		HandleTui(tuiCmd)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func HandleCheck(checkCmd *flag.FlagSet, onlyLives bool, formatCSV bool) {
	checkStreamers(onlyLives, formatCSV)
}

func HandleAdd(addCmd *flag.FlagSet, args []string) {
	// parse subcommand options (they do not exist... yet)
	addCmd.Parse(args)
	names := addCmd.Args()
	if len(names) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	addStreamer(names)
}

func HandleDel(delCmd *flag.FlagSet, args []string) {
	// parse subcommand options (they do not exist... yet)
	delCmd.Parse(args)
	names := delCmd.Args()
	if len(names) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	delStreamer(names)
}

func HandleTui(tuiCmd *flag.FlagSet) {
	err := startTUI()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
