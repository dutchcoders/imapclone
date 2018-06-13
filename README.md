# imap clone

IMAP clone will export complete imap mailboxes to Elasticsearch. 

# Configuration 

Rename `config.toml.sample` to `config.toml`.

```config.toml
elasticsearch_url="http://127.0.0.1:9200/imapclone"

[[mailbox]]
server="imap-mail.outlook.com:993"
username="username1"
password="password1"

[[mailbox]]
server="imap-mail.outlook.com:993"
username="username2"
password="password2"
```

# Run

```
go run main.go 
```
