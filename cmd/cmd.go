package cmd

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/gosuri/uilive"

	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/minio/cli"
	logging "github.com/op/go-logging"
)

var version = "0.1"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rand.Seed(time.Now().UTC().UnixNano())
}

var log = logging.MustGetLogger("cmd")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

func run(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := uilive.New()

	writer.Start()
	defer writer.Stop() // flush and stop rendering

	a := &app{
		Config: Config{},
		writer: writer,
	}

	f, err := os.Open("config.toml")
	if err != nil {
		return err
	}

	if _, err := toml.DecodeReader(f, &a); err != nil {
		return err
	}

	return a.Clone(ctx)
}

func New() *cli.App {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.DEBUG, "")

	backend1Formatter := logging.NewBackendFormatter(backend1, format)

	logging.SetBackend(backend1Formatter)

	app := cli.NewApp()
	app.Name = "imap-clone"
	app.Version = version
	app.Flags = append(app.Flags, []cli.Flag{
		cli.StringFlag{Name: "config, c", Value: "config.yaml", Usage: "Custom configuration file path"},
	}...)

	app.Action = run
	return app
}
