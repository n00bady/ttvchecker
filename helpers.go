package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func openStreamerlist() *os.File {
	streamerlist := createStreamerlist()
	f, err := os.OpenFile(streamerlist, os.O_RDONLY, 0o644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// check if the file is empty there is no point to continue
	// if there are no streamer in the file
	fi, err := f.Stat()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if fi.Size() == 0 {
		fmt.Println("The " + streamerlist + " is empty!")
		os.Exit(1)
	}

	return f
}

// checks if the streamerlist.txt exists if not
// creates a folder in users ~/.config/ called ttvchecker and
// creates a streamerlist.txt if they don't exist
// returns a string with the path to it
func createStreamerlist() string {
	configPath := constructConfigPath()

	// Create the whole path, including ~/.config when it does not exist yet.
	if err := os.MkdirAll(configPath, 0o700); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	listPath := filepath.Join(configPath, "streamerlist.txt")
	if !checkFileExist(listPath) {
		f, err := os.OpenFile(listPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			// Another process may have created it between the check and here.
			if !errors.Is(err, os.ErrExist) {
				log.Println(err)
				os.Exit(1)
			}
			return listPath
		}
		if err := f.Close(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	return listPath
}

// returns the user's .config path
// I probably could just hardcoded it but
// I feel this is a better way
func constructConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return filepath.Join(homeDir, ".config", "ttvchecker")
}

// Checks if a file exists
func checkFileExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// clears the terminal after it checks if output is a terminal
// maybe this could be in printResults.go ?
func clearTerm() {
	o, err := os.Stdout.Stat()
	if err != nil {
		log.Println(err)
	}

	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		fmt.Printf("\033[2J\033[1;1H")
	}
}

// create a GET request and return the response and an error/nil
func getResponse(link string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New("Response Status Code: " + strconv.Itoa(resp.StatusCode))
	}

	return resp, nil
}

// checks if a slice contains a value
func contains(str []string, v string) bool {
	for _, s := range str {
		if v == s {
			return true
		}
	}

	return false
}

// reads the non-empty lines of a streamer list, trimming surrounding spaces
func readStreamerNames(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			names = append(names, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

var streamerNamePattern = regexp.MustCompile(`^[a-z0-9_]{1,25}$`)

// normalizes a streamer username or twitch.tv URL, returning "" for invalid input
func normalizeStreamerName(raw string) string {
	name := strings.TrimSpace(raw)
	// net/url only populates Host when a scheme or "//" is present.
	// Supply the missing scheme for schemeless twitch.tv URLs.
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "www.twitch.tv/") || strings.HasPrefix(lower, "twitch.tv/") {
		name = "https://" + name
	}

	if u, err := neturl.Parse(name); err == nil && u.Host != "" {
		host := strings.ToLower(u.Hostname())
		if host == "twitch.tv" || host == "www.twitch.tv" {
			name = strings.Trim(u.Path, "/")
			if i := strings.Index(name, "/"); i >= 0 {
				name = name[:i]
			}
		}
	}
	name = strings.ToLower(strings.Trim(name, "/"))
	if !streamerNamePattern.MatchString(name) {
		return ""
	}

	return name
}

// -- These are used for the TUI --
// Takes the streamer string and the index, update the Rows and
// returns a tea.Cmd to show progress
func refreshStreamer(streamer string, index int) tea.Cmd {
	d := time.Duration(rand.Intn(300)+100) * time.Millisecond
	resp, _ := getResponse(url + streamer)

	if resp != nil {
		isLive, title, _ := parse(resp)
		defer resp.Body.Close()

		if isLive {
			Rows = append(Rows, table.Row{strconv.Itoa(index), streamer, "LIVE", title})
		} else {
			Rows = append(Rows, table.Row{strconv.Itoa(index), streamer, "OFFLINE", title})
		}
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return updatedMsg("done")
	})
}

// Constructs a list of all the streamers to go with the model
func initStreamerList() []string {
	names, err := readStreamerNames(createStreamerlist())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return names
}

// Takes 2 int and return the smaller one if equal returns the first one
func findMin(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}
