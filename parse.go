package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parse(response *http.Response) (bool, error) {
  islive := false
  
  doc, err := goquery.NewDocumentFromReader(response.Body)
  if err != nil {
    return false, err
  }
  doc.Find("script").Each(func(i int, s *goquery.Selection) {
    t, _ := s.Attr("type")
    if t == "application/ld+json" {
      // could this be parsed in a better way ?
      // it's json-ld
      fmt.Println(s.Text())
      if strings.Contains(s.Text(), "\"isLiveBroadcast\":true") {
        islive = true
      }
    }
  })

  return islive, nil
}
