package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
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

// LoadHistoryFile attempts to read and return the history file
// from disk and parse it as new line delimited commands
func LoadHistoryFile() ([]string, error) {
	historyBytes, err := ioutil.ReadFile(historyPath)
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(string(historyBytes), "\n")
	return lines, nil
}

// UpdateHistoryFile attempts to write a new line entry
// to the history file
func UpdateHistoryFile(line string) error {
	// If history file is empty, don't prepend \n to entry
	newline := "\n"
	historyBytes, err := ioutil.ReadFile(historyPath)
	if err != nil {
		return err
	}
	if len(historyBytes) == 0 {
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
