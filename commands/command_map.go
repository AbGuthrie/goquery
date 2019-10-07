package commands

import (
	"errors"

	prompt "github.com/c-bata/go-prompt"
)

type GoQueryCommand func(string) error

// Function Map
var CommandMap map[string]GoQueryCommand
var SuggestionsMap []prompt.Suggest

// Errors
var errArgumentError error
var errRuntimeError error

func init() {
	CommandMap = map[string]GoQueryCommand{
		".connect":    connect,
		".disconnect": disconnect,
		".query":      scheduleQuery,
		".mode":       changeMode,
		".exit":       exit,
		"ls":          listDirectory,
		"cd":          changeDirectory,
	}
	SuggestionsMap = []prompt.Suggest{
		{".connect", "Connect to a host with UUID"},
		{".disconnect", "Disconnect from a host with UUID"},
		{".query", "Schedule a query on a host"},
		{".mode", "Change print mode (json, lines, etc)"},
		{".exit", "Exit goquery"},
		{"cd", "Change directories on a remote host"},
		{"ls", "List the files in the current directory on the remote host"},
	}
	errArgumentError = errors.New("The arguments provided were incorrect for the command")
	errRuntimeError = errors.New("There was a problem executing the command")
}
