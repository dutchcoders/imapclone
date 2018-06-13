package cmd

import (
	"io"
	"net/url"

	"github.com/puerkitobio/goquery"
)

type HtmlDocument struct {
	*goquery.Document

	baseURL *url.URL
}

func NewHtmlDocumentFromReader(baseURL *url.URL, r io.Reader) (*HtmlDocument, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	return &HtmlDocument{
		Document: doc,
		baseURL:  baseURL,
	}, nil
}

func (d *HtmlDocument) ExtractLinks(elem string, attr string) []string {
	links := []string{}

	d.Find(elem).Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr(attr); !ok {
		} else {
			for _, l2 := range links {
				// only uniques
				if l2 == val {
					return
				}
			}

			links = append(links, val)
		}
	})

	return links
}
