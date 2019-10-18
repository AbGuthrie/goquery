package commands

import (
	"errors"

	prompt "github.com/c-bata/go-prompt"
)

// GoQueryCommand the signature type all command functions must conform to
type GoQueryCommand func(string) error

// CommandMap is the mapping from command line string to function call
var CommandMap map[string]GoQueryCommand

// SuggestionsMap is a prompt integration for autocomplete helpers
var SuggestionsMap []prompt.Suggest

// Errors
var errArgumentError error
var errRuntimeError error

func init() {
	CommandMap = map[string]GoQueryCommand{
		".connect":    connect,
		".disconnect": disconnect,
		".query":      query,
		".schedule":   schedule,
		".resume":     resume,
		".hosts":      printHosts,
		".history":    history,
		".clear":      clear,
		".mode":       changeMode,
		".exit":       exit,
		"ls":          listDirectory,
		"cd":          changeDirectory,
	}
	SuggestionsMap = []prompt.Suggest{
		{".connect", "Connect to a host with UUID"},
		{".disconnect", "Disconnect from a host with UUID"},
		{".query", "Schedule a query on a host and wait for results"},
		{".schedule", "Schedule a query on host but don't wait for results"},
		{".resume", "Try to fetch results for query but don't block if unavailable"},
		{".hosts", "Prints a table of connected hosts"},
		{".history", "Print the query history for a host from the current session"},
		{".clear", "Clear the terminal screen"},
		{".mode", "Change print mode (json, lines, etc)"},
		{".exit", "Exit goquery"},
		{"cd", "Change directories on a remote host"},
		{"ls", "List the files in the current directory on the remote host"},
	}
	errArgumentError = errors.New("The arguments provided were incorrect for the command")
	errRuntimeError = errors.New("There was a problem executing the command")
}
