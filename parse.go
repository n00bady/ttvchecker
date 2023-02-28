package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parse(response *http.Response) bool {
  islive := false
  
  doc, err := goquery.NewDocumentFromReader(response.Body)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  doc.Find("script").Each(func(i int, s *goquery.Selection) {
    t, _ := s.Attr("type")
    if t == "application/ld+json" {
      // could this be parsed in a better way ?
      // it's json-ld
      if strings.Contains(s.Text(), "\"isLiveBroadcast\":true") {
        islive = true
      }
    }
  })

  return islive
}
