package main

import (
	"math/rand"
	"os"
	"time"

	cmd "go.dutchsec.com/imapclone/cmd"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	app := cmd.New()
	app.Run(os.Args)
}
