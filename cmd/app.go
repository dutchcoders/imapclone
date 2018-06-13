package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	_ "log"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/gosuri/uilive"
	"github.com/gosuri/uiprogress"
)

const MAX_MEMORY = 8192 * 16

type app struct {
	Config

	writer *uilive.Writer
}

func (a *app) Clone(ctx context.Context) error {
	q := NewTaskStack()

	uiprogress.Start()

	expr := "(%IP4|%IP6)"
	expr = strings.Replace(expr, "%IP4", `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`, -1)
	expr = strings.Replace(expr, "%IP6", `(?:[0-9a-f]{1,4}\:){7}[0-9a-f]{1,4}`, -1)

	re := regexp.MustCompile(expr)

	ch := a.indexer(a.ElasticsearchURL)(ctx)
	ch = a.summary(ch)
	//	ch = a.debug(ch)

	for _, mc := range a.Mailboxes {
		var c *client.Client

		c, err := client.DialTLS(mc.Server, nil)
		if err != nil {
			return err
		}

		defer c.Logout()

		if err := c.Login(mc.Username, mc.Password); err != nil {
			log.Fatal(err)

		}

		mailboxesCh := make(chan *imap.MailboxInfo, 10)

		done := make(chan error, 1)
		go func() {
			done <- c.List("", "*", mailboxesCh)
		}()

		mailboxes := []*imap.MailboxInfo{}
		for mailbox := range mailboxesCh {
			mailboxes = append(mailboxes, mailbox)
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}

		for _, mailbox := range mailboxes {
			q.Push(&Task{
				Server:   mc.Server,
				Username: mc.Username,
				Password: mc.Password,

				FolderName: mailbox.Name,
			})
		}
	}

	type Stat struct {
		ThreadID int
		Status   string
	}

	printer := make(chan Stat)
	go func() {
		statuses := map[int]Stat{}

		for {
			s := <-printer

			statuses[s.ThreadID] = s

			for i := 0; i < len(statuses); i++ {
				status, ok := statuses[i]
				if !ok {
					continue
				}

				a.writer.Write([]byte(
					color.YellowString("%s\u001b[0K\n", status.Status),
				))
			}

			a.writer.Flush()
		}
	}()

	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer func() {
				wg.Done()
			}()

			for {
				task := q.Pop()

				if task == nil {
					fmt.Fprintln(a.writer.Bypass(), color.YellowString("Thread %d finished", threadID))
					return
				}

				func() {
					UpdateStatus := func(folderName string, status string) {
						printer <- Stat{
							ThreadID: threadID,
							Status:   fmt.Sprintf("[%d] [%s] %s", threadID, folderName, status),
						}
					}

					UpdateStatus(task.FolderName, "login")

					c, err := client.DialTLS(task.Server, nil)
					if err != nil {
						fmt.Fprintln(a.writer.Bypass(), color.RedString("Error connecting to server: %s", err.Error()))
						return
					}

					defer c.Logout()

					if err := c.Login(task.Username, task.Password); err != nil {
						fmt.Fprintln(a.writer.Bypass(), color.RedString("Error logging in: %s", err.Error()))
						return
					}

					UpdateStatus(task.FolderName, "opening folder")

					mbox, err := c.Select(task.FolderName, true)
					if err != nil {
						log.Fatal(err)
					}

					if mbox.Messages == 0 {
						return
					}

					from := uint32(1)
					to := mbox.Messages

					seqset := new(imap.SeqSet)
					seqset.AddRange(from, to)

					section := &imap.BodySectionName{}

					messages := make(chan *imap.Message, 10)

					done := make(chan error, 1)
					go func() {
						done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
					}()

					go func() {
						count := uint32(0)

						for msg := range messages {
							UpdateStatus(task.FolderName, fmt.Sprintf("downloading message %d/%d", from+count, to))

							r := msg.GetBody(section)

							m, err := mail.ReadMessage(r)
							if err != nil {
								log.Fatal(err)
							}

							doc := NewDocument()

							header := m.Header

							for key := range header {
								doc.Set(Header("headers."+strings.ToLower(key), m.Header[key]))
							}

							date := time.Now()
							// could remove rfc1123z with _2
							for _, format := range []string{time.RFC1123Z, time.RFC1123, time.RFC822, time.RFC822Z, "Mon, _2 Jan 2006 15:04:05 -0700", "Mon, _2 Jan 2006 15:04:05 -07"} {
								val, err := time.Parse(format, header.Get("Date"))
								if err != nil {
									continue
								}

								date = val
								break
							}

							doc.Store("date", date)
							doc.Store("folder", mbox.Name)

							doc.Set(AddressList("from", decodeHeader(header.Get("from"))))
							doc.Set(AddressList("to", decodeHeader(header.Get("to"))))

							doc.Store("subject", decodeHeader(header.Get("subject")))

							buff := &bytes.Buffer{}

							if data, err := ioutil.ReadAll(io.TeeReader(m.Body, buff)); err == nil {
								count := 8192
								if len(data) < count {
									count = len(data)
								}
								doc.Store("body", string(data[:count]))
							}

							m.Body = buff

							if links, err := a.extractLinks(&Message{
								m: m,
							}); err == nil {
								doc.Store("links", links)
							}

							messageID := decodeHeader(header.Get("Message-Id"))
							//messageID = messageID[1 : len(messageID)-1]
							doc.Store("message-id", messageID)

							// the last
							ips := []string{}
							for _, received := range header["Received"] {
								scanner := bufio.NewScanner(strings.NewReader(received))
								scanner.Split(bufio.ScanWords)

								scanner.Scan() // from
								scanner.Scan() // servername
								scanner.Scan() // ip
								ip := scanner.Text()

								matches := re.FindAllStringSubmatch(ip, -1)
								for _, match := range matches {
									ips = append(ips, match[0])
								}
							}

							doc.Store("remote-address", ips)

							ch <- doc

							count++
						}
					}()

					if err := <-done; err != nil {
						panic(err)
					}
				}()
			}
		}(i)
	}

	wg.Wait()

	return nil
}
