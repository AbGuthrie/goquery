package goquery

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AbGuthrie/goquery/v2/commands"
	"github.com/AbGuthrie/goquery/v2/config"
	"github.com/AbGuthrie/goquery/v2/hosts"
	"github.com/AbGuthrie/goquery/v2/models"
	"github.com/AbGuthrie/goquery/v2/utils"

	prompt "github.com/c-bata/go-prompt"
)

var apiInstance models.GoQueryAPI
var options config.Config

// Run is the entry point for a file impporting the goquery library to start the prompt REPL
func RunWithExternalCommands(api models.GoQueryAPI, _config config.Config, _externalCommandMap map[string]commands.GoQueryCommand) {
	for k, v := range _externalCommandMap {
		commands.CommandMap[k] = v
	}
	Run(api, _config)
}

// Run is the entry point for a file impporting the goquery library to start the prompt REPL
func Run(api models.GoQueryAPI, _config config.Config) {
	// Print errors/warnings with provided aliases, and print state of debug flags
	_config.Validate()

	// Set globals for executor function closure
	options = _config
	apiInstance = api

	history, err := utils.LoadHistoryFile()
	if err != nil {
		fmt.Printf("Unable to load history file %s\n", err)
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
	writeHistory := true
	defer func() {
		if !writeHistory {
			return
		}
		// Write history entry
		if err := utils.UpdateHistoryFile(input); err != nil {
			fmt.Printf("Failed to write history file: %s\n", err)
		}
	}()

	// Separate command and arguments
	input = strings.TrimSpace(input)
	args := strings.Split(input, " ")
	if len(args) == 0 {
		return
	}

	// Lookup and run command in command map
	if command, ok := commands.CommandMap[args[0]]; ok {
		err := command.Execute(apiInstance, &options, input)
		if err != nil {
			fmt.Printf("%s: %s\n", args[0], err.Error())
		}
		return
	}

	// Command not found, was this command aliased?
	alias, found := options.Aliases[args[0]]
	if !found {
		fmt.Printf("No such command: %s\n", args[0])
		return
	}
	realizedCommand, err := utils.InterpolateArguments(input, alias.Command)
	if err != nil {
		fmt.Printf("Alias error: %s\n", err)
		return
	}

	// Run the parsed and interpolated alias through executor again
	writeHistory = false
	executor(realizedCommand)
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
		// We also need to sort the final array because go traverses maps non
		// deterministically
		suggestions := make([]string, 0)

		// Add all command suggestions
		for name := range commands.CommandMap {
			suggestions = append(suggestions, name)
		}
		// Add all alias suggestions
		for name := range options.Aliases {
			suggestions = append(suggestions, name)
		}

		sort.Strings(suggestions)
		for _, suggestion := range suggestions {
			if alias, ok := options.Aliases[suggestion]; ok {
				description := alias.Description
				if len(description) == 0 {
					description = alias.Command
				}
				prompts = append(prompts, prompt.Suggest{Text: suggestion, Description: alias.Command})
			} else if command, ok := commands.CommandMap[suggestion]; ok {
				prompts = append(prompts, prompt.Suggest{Text: suggestion, Description: command.Help()})
			}
		}
		return prompt.FilterHasPrefix(prompts, command, true)
	}

	// Call into the command to ask for further suggestions
	commandStruct := commands.CommandMap[command]
	return prompt.FilterHasPrefix(commandStruct.Suggestions(in.CurrentLine()), in.GetWordBeforeCursor(), true)
}
