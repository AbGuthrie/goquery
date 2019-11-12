package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/pkg/executor"
	"github.com/go-kit/kit/log"
	"github.com/kolide/kit/logutil"
	"github.com/peterbourgon/ff"

	"github.com/AbGuthrie/goquery/commands"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"

	prompt "github.com/c-bata/go-prompt"
)

func main() {
	fs := flag.NewFlagSet("goquery", flag.ExitOnError)
	var (
		flDriver   = fs.String("driver", "", "API driver to use")
		flDebug    = fs.Bool("debug", false, "log debug information")
		flSettings = fs.String("settings", "", "settings file (optional)")
		_          = fs.String("config", "", "config file (optional)")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("GOQUERY"),
	)

	logger := logutil.NewCLILogger(*flDebug)

	// Look for missing options
	missingOpt := false
	for flag, val := range map[string]string{
		"driver": *flDriver,
	} {
		if val == "" {
			fmt.Fprintf(os.Stderr, "Missing required flag: %s\n", flag)
			missingOpt = true
		}
	}

	if missingOpt {
		os.Exit(1)
	}

	//
	// Options are setup. Time to setup
	//
	apiDriver, err := api.InitializeAPI(*flDriver)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize API: %s\n", err)
		os.Exit(1)
	}

	ex, err := executor.New(apiDriver,
		executor.WithLogger(logger),
		//executor.WithHistory(XXX),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize executor: %s\n", err)
		os.Exit(1)
	}

	// old

	history, err := utils.LoadHistoryFile()
	if err != nil {
		fmt.Printf("Unable to load history file %s\n", err)
	}

	// Initialize API integration
	if err := api.InitializeAPI(config.GetAPIDriver()); err != nil {
		log.Fatalf("Failed to initialize API driver: %s", err)
	}

	p := prompt.New(
		ex.PromptRun,
		ex.Completer,
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
		for name := range commands.CommandMap {
			commandNames = append(commandNames, name)
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
