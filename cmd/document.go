package cmd

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"golang.org/x/sync/syncmap"
)

func Header(prefix string, v []string) func(*Document) {
	return func(m *Document) {
		for i := 0; i < len(v); i++ {
			v[i] = decodeHeader(v[i])
		}
		m.Store(prefix, v)
	}
}

// Token adds the provided token into the giving Event.
func AddressList(prefix string, v string) func(*Document) {
	return func(m *Document) {
		addresses, err := mail.ParseAddressList(v)
		if err != nil {
			return
		}

		if len(addresses) == 0 {
			return
		}

		addr := addresses[0]

		if addr.Address != "" {
			m.Store(fmt.Sprintf("%s.address", prefix), addr.Address)
		}

		if addr.Name != "" {
			m.Store(fmt.Sprintf("%s.name", prefix), addr.Name)
		}
	}
}

type Document struct {
	sm *syncmap.Map
}

func (b *Document) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}

	b.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			m[keyName] = value
		}
		return true
	})

	return json.Marshal(m)
}

// New returns a new Event with the options applied.
func NewDocument(opts ...func(*Document)) *Document {
	e := &Document{
		sm: new(syncmap.Map),
	}

	e.sm.Store("date", time.Now())

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Range defines a function which ranges the underline key-values with
// the provided syncmap.
func (e *Document) Range(fx func(interface{}, interface{}) bool) {
	e.sm.Range(fx)
}

// Store adds the key and value into the event.
func (e *Document) Set(fn func(*Document)) {
	fn(e)
	// e.sm.Store(s, v)
}

// Store adds the key and value into the event.
func (e *Document) Store(s string, v interface{}) {
	e.sm.Store(s, v)
}

// Has returns true/false if the giving key exists.
func (e *Document) Has(s string) bool {
	_, ok := e.sm.Load(s)
	return ok
}

// Get retrieves a giving value for a key has string.
func (e *Document) Load(s string) (interface{}, bool) {
	return e.sm.Load(s)
}

// Get retrieves a giving value for a key has string.
func (e *Document) Get(s string) string {
	if v, ok := e.sm.Load(s); !ok {
		return ""
	} else if v, ok := v.(string); !ok {
		return ""
	} else {
		return v
	}
}
