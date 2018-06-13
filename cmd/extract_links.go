package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"github.com/mvdan/xurls"

	"github.com/fatih/color"
)

func (a *app) extractLinks(part Part) ([]string, error) {
	links := []string{}

	r := io.TeeReader(part.NewReader(), ioutil.Discard)

	if mediaType, params, err := part.MediaType(); err != nil {
		fmt.Fprintln(a.writer.Bypass(), color.RedString("Error retrieving mediaType: %s", err.Error()))
	} else if strings.HasPrefix(mediaType, "multipart") {
		pr := multipart.NewReader(r, params["boundary"])
		for {
			part2, err := pr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Fprintln(a.writer.Bypass(), color.RedString("Error retrieving nextpart(01): %s: %#+v", err.Error(), part))
				break
			}

			if newLinks, err := a.extractLinks(&MessagePart{
				m: part2,
			}); err != nil {
				fmt.Fprintln(a.writer.Bypass(), color.RedString("Error retrieving nextpart(02): %s", err.Error()))
			} else {
				links = uniqueAppend(links, newLinks...)
			}
		}
	} else if mediaType == "text/html" {
		if hd, err := NewHtmlDocumentFromReader(nil, r); err == io.EOF {
		} else if err != nil {
			fmt.Fprintln(a.writer.Bypass(), color.RedString("Error parsing html part: %s", err.Error()))
		} else {
			newLinks := hd.ExtractLinks("a", "href")
			links = uniqueAppend(links, newLinks...)
		}
	} else {
		buff, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Fprintln(a.writer.Bypass(), color.RedString("Error parsing html part: %s", err.Error()))
		}

		newLinks := xurls.Strict.FindAllString(string(buff), -1)
		links = uniqueAppend(links, newLinks...)
	}

	return links, nil
}
