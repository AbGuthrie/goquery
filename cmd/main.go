package main

import (
	"fmt"
	"log"
	"sort"
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
		subPrefix = " | " + currentHost.ComputerName + ":" + currentHost.CurrentDirectory
	}
	return fmt.Sprintf("goquery%s> ", subPrefix), true
}

func executor(input string) {
	args := strings.Split(input, " ") // Separate command and arguments
	if command, ok := commands.CommandMap[args[0]]; ok {
		err := command.Execute(input)
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
	command := strings.Split(in.CurrentLine(), " ")[0]
	// Nothing has been typed at the prompt
	if command == "" {
		return []prompt.Suggest{}
	}

	// Suggest any top level command
	if _, ok := commands.CommandMap[command]; !ok {
		prompts := []prompt.Suggest{}
		// We also need to sort the command because go traverses maps non
		// deterministically
		commandNames := make([]string, 0)
		for k, _ := range commands.CommandMap {
			commandNames = append(commandNames, k)
		}
		sort.Strings(commandNames)
		for _, commandName := range commandNames {
			prompts = append(prompts, prompt.Suggest{commandName, commands.CommandMap[commandName].Help()})
		}
		return prompt.FilterHasPrefix(prompts, command, true)
	}

	// Call into the command to ask for further suggestions
	commandStruct := commands.CommandMap[command]
	return prompt.FilterHasPrefix(commandStruct.Suggestions(in.CurrentLine()), in.GetWordBeforeCursor(), true)
}
