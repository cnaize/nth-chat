package term

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

var cmdSuggests = []prompt.Suggest{
	{Text: "send", Description: "Send message"},
	{Text: "pull", Description: "Pull messages"},
	{Text: "login", Description: "Login into account"},
	{Text: "register", Description: "Register an account"},
	{Text: "quit", Description: "Quit the app"},
}

func Completer(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.CurrentLineBeforeCursor(), " ")
	if len(args) < 1 {
		return []prompt.Suggest{}
	}

	cmd := args[0]
	switch cmd {
	case "login", "register":
		if len(args) == 2 {
			return []prompt.Suggest{{Text: "username", Description: "Your username"}}
		} else if len(args) == 3 {
			return []prompt.Suggest{{Text: "password", Description: "Your password"}}
		} else {
			return []prompt.Suggest{}
		}
	case "send":
		if len(args) == 2 {
			return []prompt.Suggest{{Text: "username", Description: "Message recipient (empty for public)"}}
		} else if len(args) > 2 {
			return []prompt.Suggest{{Text: "text", Description: "Message text"}}
		}
	}

	if suggests := prompt.FilterHasPrefix(cmdSuggests, cmd, true); len(suggests) > 0 {
		return suggests
	}

	return cmdSuggests
}
