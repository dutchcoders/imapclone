package cmd

import _ "log"

func (b *app) summary(ch chan *Document) chan *Document {
	out := make(chan *Document)

	go func() {
		for doc := range out {
			ch <- doc
		}

		close(out)
	}()

	return out
}
