package main

import (
	"encoding/json"
	"net/http"

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

	// These nested structs are annoying...
	// I added some fields that I don't use right now, maybe for the future ???
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var data struct {
			Context string `json:"@context"`
			Graph   []struct {
				Type        string `json:"@type"`
				Description string `json:"description"`
				Name        string `json:"name"`
				EmbedURL    string `json:"embedUrl"`
				Publication struct {
					Type             string `json:"@type"`
					IsLiveBroadcast  bool   `json:"isLiveBroadcast"`
					StartDate        string `json:"startDate"`
					EndDate          string `json:"endDate"`
				} `json:"publication"`
				ThumbnailURL []string `json:"thumbnailUrl"`
				UploadDate   string   `json:"uploadDate"`
			} `json:"@graph"`
		}

		if err := json.Unmarshal([]byte(s.Text()), &data); err != nil {
			return
		}

		for _, item := range data.Graph {
			if item.Type == "VideoObject" {
				islive = item.Publication.IsLiveBroadcast
				title = item.Description
				break
			}
		}
	})

	return islive, title, nil
}
