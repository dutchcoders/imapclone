package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"

	"golang.org/x/net/html/charset"
)

type Part interface {
	MediaType() (string, map[string]string, error)
	NewReader() io.Reader
}

type MessagePart struct {
	m *multipart.Part
}

func (m *MessagePart) MediaType() (string, map[string]string, error) {
	contentType := m.m.Header.Get("Content-Type")

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", nil, err
	}

	return mediaType, params, nil
}

func (m *MessagePart) NewReader() io.Reader {
	var body io.Reader = m.m

	// transfer encoding, should we split the transfer encoding and the content type?
	contentTransferEncoding := ""
	if v, ok := m.m.Header["Content-Transfer-Encoding"]; ok {
		contentTransferEncoding = v[0]
	}

	if contentTransferEncoding == "base64" {
		body = base64.NewDecoder(base64.StdEncoding, body)
	} else if contentTransferEncoding == "quoted-printable" {
		body = quotedprintable.NewReader(body)
	} else if false /* iso */ {
		// https://github.com/gdamore/encoding/
	} else {
	}

	return body
}

type Message struct {
	m *mail.Message
}

func (m *Message) MediaType() (string, map[string]string, error) {
	contentType := m.m.Header.Get("Content-Type")

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", nil, err
	}

	return mediaType, params, nil
}

func (m *Message) NewReader() io.Reader {
	var body io.Reader = m.m.Body

	// transfer encoding, should we split the transfer encoding and the content type?
	contentTransferEncoding := ""
	if v, ok := m.m.Header["Content-Transfer-Encoding"]; ok {
		contentTransferEncoding = v[0]
	}

	if contentTransferEncoding == "base64" {
		body = base64.NewDecoder(base64.StdEncoding, body)
	} else if contentTransferEncoding == "quoted-printable" {
		body = quotedprintable.NewReader(body)
	} else if false /* iso */ {
		// https://github.com/gdamore/encoding/
	} else {
	}

	return body
}

func uniqueAppend(a []string, val ...string) []string {
	b := a

	for _, l3 := range val {
		found := false

		for _, l2 := range b {
			// only uniques
			if l2 == l3 {
				found = true
				break
			}
		}

		if !found {
			b = append(b, l3)
		}
	}

	return b

}

// RFC 1342: Non-ASCII Mail Headers
func decodeHeader(str string) string {
	dec := new(mime.WordDecoder)

	dec.CharsetReader = func(cs string, input io.Reader) (io.Reader, error) {
		return charset.NewReaderLabel(cs, input)
	}

	val, err := dec.DecodeHeader(str)
	if err != nil {
		fmt.Println(err.Error(), val)
		return str
	}

	return val
}
