package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
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
				results = append(results, stream{name: streamer, live: isLive, title: title, url: url + streamer})
			} else if !onlyLives {
				results = append(results, stream{name: streamer, live: isLive, title: title, url: url + streamer})
			}
		} else {
			// show the stream as offline if the response is not what you expect
			results = append(results, stream{name: streamer, live: false, title: "", url: url + streamer})
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

// Adds normalized, unique streamers to the config file
func addStreamer(names []string) {
	streamerlist := createStreamerlist()
	existing, err := readStreamerNames(streamerlist)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	seen := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		if normalized := normalizeStreamerName(name); normalized != "" {
			seen[normalized] = struct{}{}
		}
	}

	var added []string
	for _, raw := range names {
		name := normalizeStreamerName(raw)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		added = append(added, name)
	}
	if len(added) == 0 {
		fmt.Println("No new streamers added.")
		return
	}

	f, err := os.OpenFile(streamerlist, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	for _, name := range added {
		if _, err := fmt.Fprintln(f, name); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
	fmt.Println(added, "added.")
}

// Deletes streamers from the config file, matching usernames case-insensitively
func delStreamer(names []string) {
	streamerlist := createStreamerlist()
	existing, err := readStreamerNames(streamerlist)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	targets := make(map[string]struct{}, len(names))
	for _, raw := range names {
		if name := normalizeStreamerName(raw); name != "" {
			targets[name] = struct{}{}
		}
	}

	var kept, deleted []string
	for _, line := range existing {
		if _, ok := targets[strings.ToLower(strings.TrimSpace(line))]; ok {
			deleted = append(deleted, line)
			continue
		}
		kept = append(kept, line)
	}
	if len(deleted) == 0 {
		fmt.Println("No streamers deleted.")
		return
	}

	// Write the new list to a temporary file in /tmp (via os.TempDir) and
	// rename it into place, so a failed write cannot leave a truncated list.
	tmp, err := os.CreateTemp("", ".streamerlist-*")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	for _, line := range kept {
		if _, err := fmt.Fprintln(tmp, line); err != nil {
			tmp.Close()
			log.Println(err)
			os.Exit(1)
		}
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		log.Println(err)
		os.Exit(1)
	}
	if err := tmp.Close(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := os.Rename(tmpName, streamerlist); err != nil {
		// /tmp may be on a different filesystem from the config directory,
		// where rename cannot work. Copy the fully prepared contents instead.
		if !errors.Is(err, syscall.EXDEV) {
			log.Println(err)
			os.Exit(1)
		}
		src, err := os.Open(tmpName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer src.Close()
		dst, err := os.OpenFile(streamerlist, os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			log.Println(err)
			os.Exit(1)
		}
		if err := dst.Close(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println(deleted, "deleted.")
}
