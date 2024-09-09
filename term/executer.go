package term

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cnaize/nth-chat/client"
	"github.com/cnaize/nth-chat/config"
	"github.com/cnaize/nth-chat/entity"
)

type ExecuterHandleFn func(s string, args []string) error

type Executor struct {
	cfg      *config.Config
	client   *client.Client
	hanlders map[string]ExecuterHandleFn
}

func NewExecuter(cfg *config.Config, client *client.Client) *Executor {
	e := &Executor{
		cfg:    cfg,
		client: client,
	}

	e.hanlders = map[string]ExecuterHandleFn{
		"send":     e.handleSend,
		"pull":     e.handlePull,
		"quit":     e.handleQuit,
		"login":    e.handleLogin,
		"register": e.handleRegister,
	}

	return e
}

func (e *Executor) Handle(s string) {
	args := strings.Split(strings.Trim(s, " "), " ")
	if len(args) < 1 {
		return
	}

	cmd := args[0]
	// check auth
	if (cmd != "quit" && cmd != "login" && cmd != "register") &&
		(e.cfg.Creds.Username == "" || e.cfg.Creds.Password == "") {
		fmt.Println("You have to login first")
		return
	}

	hanlder, ok := e.hanlders[cmd]
	if !ok {
		return
	}

	if err := hanlder(s, args); err != nil {
		fmt.Println("Error:", err)
	}
}

func (e *Executor) handleQuit(s string, args []string) error {
	fmt.Println(e.client.Close())
	fmt.Println("Bye!")
	os.Exit(0)

	return nil
}

func (e *Executor) handleRegister(s string, args []string) error {
	if len(args) != 3 {
		return entity.ErrInvalidArguments
	}

	username := args[1]
	password := args[2]
	if username == "" || password == "" {
		return entity.ErrEmptyCredentials
	}

	panic("implement me")
}

func (e *Executor) handleLogin(s string, args []string) error {
	if len(args) != 3 {
		return entity.ErrInvalidArguments
	}

	username := args[1]
	password := args[2]
	if username == "" || password == "" {
		return entity.ErrEmptyCredentials
	}

	e.cfg.Creds.Username = username
	e.cfg.Creds.Password = password
	fmt.Println("Success: logged in as", e.cfg.Creds.Username)

	// pull messages
	return e.handlePull("pull", []string{"pull"})
}

func (e *Executor) handleSend(s string, args []string) error {
	if len(args) < 3 {
		return entity.ErrInvalidArguments
	}

	if err := e.client.SendMessage(
		entity.Message{
			From: e.cfg.Creds.Username,
			To:   args[1],
			Text: strings.Join(args[2:], " "),
		},
	); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (e *Executor) handlePull(s string, args []string) error {
	messages, err := e.client.PullMessages()
	if err != nil {
		return fmt.Errorf("pull messages: %w", err)
	}

	for _, msg := range messages {
		fmt.Printf("%s %s -> %s: %s\n", time.Now().Local().Format(time.DateTime), msg.From, msg.To, msg.Text)
	}

	return nil
}
