package main

import (
	"flag"
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/cnaize/nth-chat/client"
	"github.com/cnaize/nth-chat/config"
	"github.com/cnaize/nth-chat/term"
	"github.com/nats-io/nats.go"
)

var cfg config.Config

func run() error {
	flag.StringVar(&cfg.ServerURL, "server-url", nats.DefaultURL, "nats server url")
	flag.Parse()

	client, err := client.NewClient(&cfg)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}
	defer func() {
		fmt.Println(client.Close())
		fmt.Println("Bye!")
	}()

	p := prompt.New(
		term.NewExecuter(&cfg, client).Handle,
		term.Completer,
		prompt.OptionTitle("pull based chat"),
		prompt.OptionPrefix("nth> "),
		prompt.OptionInputTextColor(prompt.Yellow),
	)
	p.Run()

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
