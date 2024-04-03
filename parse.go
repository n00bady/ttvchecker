package main

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Takes an http response and return a bool and an error
// First I parse the html response and search for the
// <script> token that has type = application/ls+json
// this one only exist on streams that are live
// and include inside various info about the stream
// such as the "isLiveBroadcast":true which if it exists
// it's obviously an Live stream
// I imagine there is better way to parse the json-ld ???
func parse(response *http.Response) (bool, string, error) {
	islive := false
    title := ""

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return false, title, err
	}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		t, _ := s.Attr("type")
		if t == "application/ld+json" {
			// could this be parsed in a better way ?
			// it's json-ld
			if strings.Contains(s.Text(), "\"isLiveBroadcast\":true") {
				islive = true
			}
            // find the title if it exists and take only the substring
            // that we care about
            if strings.Contains(s.Text(), "\"description\":") {
                i := strings.Index(s.Text(), "\"description\":")
                j := strings.Index(s.Text(), "\"embedUrl\":")

                title = s.Text()[i+15:j-2]
            }
		}
	})

	return islive, title, nil
}
