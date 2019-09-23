package main

import (
	"fmt"
	"github.com/AbGuthrie/goquery/commands"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

// TODO pull this from commands package
var suggestions = []prompt.Suggest{
	{".connect", "Connect to a host with UUID"},
	{".exit", "Exit goquery"},
}

func livePrefix() (string, bool) {
	// TODO evaluate some state and return true if we want the
	// preview pane dropdown
	return "", false
}

func executor(input string) {
	args := strings.Split(input, " ") // Separate command and arguments
	if command, ok := commands.CommandMap[args[0]]; ok {
		err := command(input)
		if err != nil {
			fmt.Printf("%s: %s!\n", args[0], err.Error())
		}
	} else {
		fmt.Printf("No such command: %s\n", args[0])
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("goquery> "),
		prompt.OptionLivePrefix(livePrefix),
		prompt.OptionTitle("goquery"),
	)
	p.Run()
}
