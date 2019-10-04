package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/AbGuthrie/goquery/commands"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"

	prompt "github.com/c-bata/go-prompt"
)

func main() {
	history, err := utils.LoadHistoryFile()
	if err != nil {
		log.Fatal(err)
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("goquery> "),
		prompt.OptionLivePrefix(refreshLivePrefix),
		prompt.OptionTitle("goquery"),
		prompt.OptionHistory(history),
	)
	p.Run()
}

func refreshLivePrefix() (string, bool) {
	// Prototype for showing current connected host state in
	// input line prefix
	subPrefix := ""
	currentHost, err := hosts.GetCurrentHost()
	if err == nil {
		subPrefix = " | " + currentHost.UUID + ":" + currentHost.CurrentDirectory
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
	if err := utils.UpdateHistoryFile(input); err != nil {
		fmt.Printf("%s\n", err)
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(commands.SuggestionsMap, w, true)
}
