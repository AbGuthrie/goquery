package main

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/commands"
	"github.com/AbGuthrie/goquery/hosts"

	prompt "github.com/c-bata/go-prompt"
)

func refreshLivePrefix() (string, bool) {
	// Prototype for showing current connected host state in
	// input line prefix
	subPrefix := ""
	currentHost, err := hosts.GetCurrentHost()
	if err == nil {
		subPrefix = " | " + currentHost.ComputerName + ":" + currentHost.CurrentDirectory
	}
	return fmt.Sprintf("goquery%s> ", subPrefix), true
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
	return prompt.FilterHasPrefix(commands.SuggestionsMap, w, true)
}

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("goquery> "),
		prompt.OptionLivePrefix(refreshLivePrefix),
		prompt.OptionTitle("goquery"),
	)
	p.Run()
}
