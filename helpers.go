package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func openStreamerlist() *os.File {
	streamerlist := createStreamerlist()
	f, err := os.OpenFile(streamerlist, os.O_RDONLY, 644)
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

	// if the config folder doesn't exits then create it and the streamerlist.txt
	if !checkFileExist(configPath) {
		err := os.Mkdir(configPath, os.ModePerm)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		f, err := os.Create(configPath + "/streamerlist.txt")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		f.Close()
	}

	// if the config folder does exist but the streamerlist.txt does not then create it
	if !checkFileExist(configPath + "/streamerlist.txt") {
		_, err := os.Create(configPath + "/streamerlist.txt")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	return configPath + "/streamerlist.txt"
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
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
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
