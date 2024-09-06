package term

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cnaize/nth-chat/config"
	"github.com/cnaize/nth-chat/entity"
	"github.com/cnaize/nth-chat/client"
)

func Executor(cfg *config.Config, client *client.Client) func(s string) {
	return func(s string) {
		args := strings.Split(strings.Trim(s, " "), " ")
		if len(args) < 1 {
			return
		}

		cmd := args[0]
		// check auth
		if (cmd != "quit" && cmd != "signin" && cmd != "signup") &&
			(cfg.Username == "" || cfg.Password == "") {
			fmt.Println("You have to sign in first")
			return
		}

		switch cmd {
		case "quit":
			fmt.Println(client.Close())
			fmt.Println("Bye!")
			os.Exit(0)
		case "signup":
			if len(args) != 3 {
				fmt.Println("Error: invalid arguments")
				return
			}

			username := args[1]
			password := args[2]
			if username == "" || password == "" {
				fmt.Println("Error: empty username or password")
				return
			}

			panic("implement me")
		case "signin":
			if len(args) != 3 {
				fmt.Println("Error: invalid arguments")
				return
			}

			username := args[1]
			password := args[2]
			if username == "" || password == "" {
				fmt.Println("Error: empty username or password")
				return
			}

			cfg.Username = username
			cfg.Password = password
			fmt.Println("Success: signed in as", cfg.Username)

			// pull messages
			Executor(cfg, client)("pull")
		case "send":
			if len(args) < 2 {
				return
			}

			if err := client.SendMessage(
				entity.Message{
					From: cfg.Username,
					To:   args[1],
					Text: strings.Join(args[2:], " "),
				},
			); err != nil {
				fmt.Println("Error: send message:", err)
			}
		case "pull":
			messages, err := client.PullMessages()
			if err != nil {
				fmt.Println("Error: pull messages:", err)
				return
			}

			for _, msg := range messages {
				fmt.Printf("%s %s -> %s: %s\n", time.Now().Local().Format(time.DateTime), msg.From, msg.To, msg.Text)
			}
		}
	}
}
