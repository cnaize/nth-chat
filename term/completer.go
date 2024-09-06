package term

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

var commands = []prompt.Suggest{
	{Text: "send", Description: "Send message"},
	{Text: "pull", Description: "Pull messages"},
	{Text: "quit", Description: "Quit the app"},
	{Text: "signin", Description: "Sing in account"},
	{Text: "signup", Description: "Sign up account"},
}

func Completer(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.CurrentLineBeforeCursor(), " ")
	if len(args) < 1 {
		return []prompt.Suggest{}
	}

	cmd := args[0]
	switch cmd {
	case "signup", "signin":
		if len(args) == 2 {
			return []prompt.Suggest{{Text: "username", Description: "Your username"}}
		} else if len(args) == 3 {
			return []prompt.Suggest{{Text: "password", Description: "Your password"}}
		} else {
			return []prompt.Suggest{}
		}
	case "send":
		if len(args) == 2 {
			return []prompt.Suggest{{Text: "username", Description: "Message recipient"}}
		} else if len(args) > 2 {
			return []prompt.Suggest{{Text: "text", Description: "Message text"}}
		}
	}

	if sugs := prompt.FilterHasPrefix(commands, cmd, true); len(sugs) > 0 {
		return sugs
	}

	return commands
}
