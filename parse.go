package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Takes an http response and return a bool and an error
// First I parse the html response and search for the
// <script> token that has type = application/ld+json
// then partially create the data structure for the json
// with only the fields I am interested on and use encoding/json
// to Unmarshal them.
func parse(response *http.Response) (bool, string, error) {
	islive := false
	var title string

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return false, title, err
	}

	// Only decode the fields needed to decide whether the stream is live.
	doc.Find("script[type='application/ld+json']").EachWithBreak(func(i int, s *goquery.Selection) bool {
		var data struct {
			Graph []struct {
				Type        string `json:"@type"`
				Description string `json:"description"`
				Publication struct {
					IsLiveBroadcast bool `json:"isLiveBroadcast"`
				} `json:"publication"`
			} `json:"@graph"`
		}

		// A script can be malformed or describe something else entirely; keep
		// looking in the remaining blocks in that case.
		if err := json.Unmarshal([]byte(s.Text()), &data); err != nil {
			return true
		}

		for _, item := range data.Graph {
			if item.Type == "VideoObject" {
				islive = item.Publication.IsLiveBroadcast
				title = strings.TrimSpace(item.Description)
				// The first VideoObject is the current stream; later blocks
				// can contain stale cached data and must not overwrite it.
				return false
			}
		}
		return true
	})

	return islive, title, nil
}
