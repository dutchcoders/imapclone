package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

func (b *app) indexer(dst string) func(ctx context.Context) chan *Document {
	index := "imapclone"

	u, err := url.Parse(dst)
	if err != nil {
		panic(err)
	}

	if len(u.Path) > 2 {
		index = u.Path[1:]
	}

	u.Path = ""

	elasticsearchURL := u.String()

	return func(ctx context.Context) chan *Document {
		ch := make(chan *Document, 1000)

		es, err := elastic.NewClient(
			elastic.SetHttpClient(&http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 5,
					TLSClientConfig:     &tls.Config{},
				},
				Timeout: time.Duration(20) * time.Second,
			}),
			elastic.SetSniff(false),
			elastic.SetURL(elasticsearchURL),
		)
		if err != nil {
			panic(err)
		}

		bulk := es.Bulk()
		go func() {
			count := 0

			do := func() {
				if bulk.NumberOfActions() == 0 {
				} else if response, err := bulk.Do(context.Background()); err != nil {
					fmt.Printf("Error indexing: %s\n", err.Error())
				} else {
					indexed := response.Indexed()
					count += len(indexed)

					for _, item := range response.Failed() {
						fmt.Printf("Error indexing item: %s with error: %+v\n", item.Id, *item.Error)
					}

					// fmt.Fprintf(b.writer, color.BlueString("[ ] Bulk indexing: %d total %d\n", len(indexed), count))
					// b.writer.Flush()
				}
			}

			defer do()

			for {
				select {
				case doc := <-ch:
					bulk = bulk.Add(elastic.NewBulkIndexRequest().
						Index(index).
						Type("message").
						Id(doc.Get("messageid")).
						Doc(doc),
					)

					if bulk.NumberOfActions() < 100 {
						continue
					}
				case <-ctx.Done():
					return

				case <-time.After(time.Second * 10):
				}

				do()
			}
		}()

		return ch
	}
}
