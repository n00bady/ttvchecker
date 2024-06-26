package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	url       string = "https://www.twitch.tv/"
	ISLIVE    string = "\"isLiveBroadcast\":true"
	userAgent        = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0"
)

type stream struct {
	name  string
	live  bool
	title string
	url   string
}

// Checks the state of all the streamers on the config file
// and prints a table with that state
func checkStreamers(onlyLives bool, formatCSV bool) (streams []stream) {
	f := openStreamerlist()

	// take each line in the file and make a GET http request
	// then parse the response and figure out if the stream is live or not.
	var results []stream
	fScanner := bufio.NewScanner(f)
	fScanner.Split(bufio.ScanLines)
	for fScanner.Scan() {
		streamer := fScanner.Text()
		fmt.Println("Checking ", yellow+streamer+reset, "...")

		resp, err := getResponse(url + streamer)
		if err != nil {
			log.Println(err)
		}

		if resp != nil {
			isLive, title, err := parse(resp)
			if err != nil {
				log.Println(err)
			}
			defer resp.Body.Close()

			// check if the onlyLives option is enabled and add only the
			// streams that are live on the results if it is
			if onlyLives && isLive {
				results = append(results, stream{name: streamer, live: isLive, title: title, url: url+streamer})
			} else if !onlyLives {
				results = append(results, stream{name: streamer, live: isLive, title: title, url: url+streamer})
			}
		} else {
			// show the stream as offline if the response is not what you expect
			results = append(results, stream{name: streamer, live: false, title: "", url: url+streamer})
		}
		// add a delay between each request so we won't get banned :S
		time.Sleep(500 * time.Millisecond)
	}
	defer f.Close() // don't forget to close it

	// print the results in a csv format if the the option is enabled
	if formatCSV {
		clearTerm()
		csvPrint(results)

		return nil
	}

	// print the results as a table in stdout
	clearTerm()
	pPrint(results)
	fmt.Println()

	return nil
}

// Adds a streamer to the config file
func addStreamer(name []string) {
	streamerlist := createStreamerlist()

	f, err := os.OpenFile(streamerlist, os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	for _, n := range name {
		if _, err := f.WriteString(n + "\n"); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
	fmt.Println(name, " added.")
}

// Deletes a streamer from the config file
func delStreamer(name []string) {
	var tmp []string
	streamerlist := createStreamerlist()

	f, err := os.OpenFile(streamerlist, os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	// is this the correct way to do find and replace/delete ?
	fScanner := bufio.NewScanner(f)
	fScanner.Split(bufio.ScanLines)
	for fScanner.Scan() {
		line := fScanner.Text()
		if !contains(name, line) {
			tmp = append(tmp, fScanner.Text())
		}
	}

	f.Seek(0, io.SeekStart)
	if err := os.Truncate(streamerlist, 0); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	buf := bufio.NewWriter(f)
	for _, v := range tmp {
		_, err := buf.WriteString(v + "\n")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
	buf.Flush()

	fmt.Println(name, " deleted.")
}
