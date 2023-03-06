package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const url string = "https://www.twitch.tv/"
const ISLIVE string = "\"isLiveBroadcast\":true"

type stream struct { 
  name string
  live bool
  link string
}

// Checks the state of all the streamers on the config file
// and prints a table with that state
func checkStreamers() (streams []stream) {

  streamerlist := createStreamerlist()
  f, err := os.OpenFile(streamerlist, os.O_RDONLY, 644)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  defer f.Close()

  // check if the file is empty there is no point to continue
  // if there are no streamer in the file
  fi, err := f.Stat()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  if fi.Size() == 0 {
    fmt.Println("The "+streamerlist+" is empty!")
    os.Exit(1)
  }

  // take each line by the file and make a get http request
  var results []stream
  fScanner := bufio.NewScanner(f)
  fScanner.Split(bufio.ScanLines)
  for fScanner.Scan() {
    streamer := fScanner.Text()
    resp, err := http.Get(url+streamer)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
      fmt.Println("Response Status Code: ", resp.StatusCode)
      os.Exit(1)
    }

    isLive, err := parse(resp)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    results = append(results, stream{name: streamer, live: isLive, link: url+streamer})

    // add a delay between each request so we won't get banned :S
    time.Sleep(3 * time.Second)
  }
  pPrint(results)

  return nil
}

// Adds a streamer to the config file
func addStreamer(name string) {

  streamerlist := createStreamerlist()
  
  f, err := os.OpenFile(streamerlist, os.O_APPEND|os.O_WRONLY, os.ModePerm)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  defer f.Close()

  if _, err := f.WriteString(name+"\n"); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  fmt.Println(name, " added.")
}

// Deletes a streamer from the config file
func delStreamer(name string) {

  var tmp []string
  streamerlist := createStreamerlist()

  f, err := os.OpenFile(streamerlist, os.O_RDWR, 0644)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  defer f.Close()

  // is this the correct way to do find and replace/delete ?
  fScanner := bufio.NewScanner(f)
  fScanner.Split(bufio.ScanLines)
  for fScanner.Scan() {
    line := fScanner.Text()
    if name != line {
      tmp = append(tmp, fScanner.Text()) 
    }
  }
  f.Seek(0, io.SeekStart)
  if err := os.Truncate(streamerlist, 0); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  buf := bufio.NewWriter(f)
  for _, v := range tmp {
    _, err := buf.WriteString(v + "\n")
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  }
  buf.Flush()

  fmt.Println(name, " deleted.")
}

func createStreamerlist() string {

  configPath := constructConfigPath()

  // if the config folder doesn't exits then create it and the streamerlist.txt
  if !checkFileExist(configPath) {
    err := os.Mkdir(configPath, os.ModePerm)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    f, err := os.Create(configPath+"/streamerlist.txt")
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    f.Close()
  }

  // if the config folder does exist but the streamerlist.txt does not then create it
  if !checkFileExist(configPath+"/streamerlist.txt") {
    _, err := os.Create(configPath+"/streamerlist.txt")
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  }

  return configPath+"/streamerlist.txt"
}

func constructConfigPath() string {

  homeDir, err := os.UserHomeDir()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  return filepath.Join(homeDir, ".config", "ttvchecker")
}

func checkFileExist(path string) bool {

  if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
    return false
  }

  return true
}
