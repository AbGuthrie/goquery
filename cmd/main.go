package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/AbGuthrie/goquery/commands"
	"github.com/AbGuthrie/goquery/hosts"

	prompt "github.com/c-bata/go-prompt"
)

var historyPath string

func init() {
	// Populate historyPath and create history file if needed
	usr, err := user.Current()
	if err != nil {
		log.Fatal(fmt.Printf("Failed to fetch user info for home directory: %s", err))
	}
	goQueryPath := path.Join(usr.HomeDir, ".goquery")
	historyPath = path.Join(goQueryPath, "history")

	// Create directory and file home if it doesn't exist yet
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		os.MkdirAll(goQueryPath, os.ModePerm)
		emptyHistory, err := os.Create(historyPath)
		emptyHistory.Close()
		if err != nil {
			log.Fatal(fmt.Printf("Failed to create history file: %s", err))
		}
	}
}

func main() {
	history, err := loadHistoryFile()
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

// Begin history file functions
// ----------------------------

func loadHistoryFile() ([]string, error) {
	historyBytes, err := ioutil.ReadFile(historyPath)
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(string(historyBytes), "\n")
	return lines, nil
}

func updateHistoryFile(line string) error {
	// If history file is empty, don't prepend \n to entry
	newline := "\n"
	historyBytes, err := ioutil.ReadFile(historyPath)
	if err != nil {
		return err
	}
	if string(historyBytes) == "" {
		newline = ""
	}
	// Write line entry to history file
	historyFile, err := os.OpenFile(historyPath, os.O_APPEND|os.O_WRONLY, 0644)
	defer historyFile.Close()
	if err != nil {
		return err
	}
	if _, err := historyFile.Write([]byte(newline + line)); err != nil {
		return err
	}
	return nil
}

// begin go-prompt integration functions
// ------------------------------------

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
	if err := updateHistoryFile(input); err != nil {
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
