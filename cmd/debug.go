package cmd

import (
	"fmt"
	_ "log"
)

func (b *app) debug(ch chan *Document) chan *Document {
	out := make(chan *Document)

	go func() {
		for doc := range out {
			fmt.Println(doc.Get("subject"))
			ch <- doc
		}

		close(out)
	}()

	return out
}
