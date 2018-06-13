package cmd

import _ "log"

type Config struct {
	ElasticsearchURL string `toml:"elasticsearch_url"`

	Mailboxes []MailboxConfig `toml:"mailbox"`
}

type MailboxConfig struct {
	Server string `toml:"server"`

	Username string `toml:"username"`
	Password string `toml:"password"`
}
