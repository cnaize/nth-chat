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
		(e.cfg.Credentials.Username == "" || e.cfg.Credentials.Password == "") {
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
	if len(e.cfg.Credentials.Username) > 0 {
		fmt.Println("Already logged in as:", e.cfg.Credentials.Username)
		return nil
	}

	if len(args) != 3 {
		return entity.ErrInvalidArguments
	}

	username := args[1]
	password := args[2]
	if username == "" || password == "" {
		return entity.ErrEmptyCredentials
	}

	e.cfg.Credentials.Username = username
	e.cfg.Credentials.Password = password

	fmt.Println("Success: logged in as:", e.cfg.Credentials.Username)

	return nil
}

func (e *Executor) handleSend(s string, args []string) error {
	if len(args) < 3 {
		return entity.ErrInvalidArguments
	}

	if err := e.client.SendMessage(
		entity.Message{
			From: e.cfg.Credentials.Username,
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
		to := msg.To
		if to == "" {
			to = "public"
		}
		fmt.Printf("%s %s -> %s: %s\n", time.Now().Local().Format(time.DateTime), msg.From, to, msg.Text)
	}

	return nil
}
