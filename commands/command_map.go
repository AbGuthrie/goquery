package commands

import (
	"errors"

	prompt "github.com/c-bata/go-prompt"
)

// GoQueryCommand defines the functions required to add a new command to goquery
type GoQueryCommand struct {
	Execute     func(string) error
	Help        func() string
	Suggestions func(string) []prompt.Suggest
}

// CommandMap is the mapping from command line string to GoQueryCommand
// structure
var CommandMap map[string]GoQueryCommand

// Errors
var errArgumentError error
var errRuntimeError error

func init() {
	CommandMap = map[string]GoQueryCommand{
		".connect":    GoQueryCommand{connect, connectHelp, connectSuggest},
		".disconnect": GoQueryCommand{disconnect, disconnectHelp, disconnectSuggest},
		".query":      GoQueryCommand{query, queryHelp, querySuggest},
		".schedule":   GoQueryCommand{schedule, scheduleHelp, scheduleSuggest},
		".resume":     GoQueryCommand{resume, resumeHelp, resumeSuggest},
		".hosts":      GoQueryCommand{printHosts, printHostsHelp, printHostsSuggest},
		".history":    GoQueryCommand{history, historyHelp, historySuggest},
		".clear":      GoQueryCommand{clear, clearHelp, clearSuggest},
		".mode":       GoQueryCommand{changeMode, changeModeHelp, changeModeSuggest},
		".exit":       GoQueryCommand{exit, exitHelp, exitSuggest},
		"ls":          GoQueryCommand{listDirectory, listDirectoryHelp, listDirectorySuggest},
		"cd":          GoQueryCommand{changeDirectory, changeDirectoryHelp, changeDirectorySuggest},
	}
	errArgumentError = errors.New("The arguments provided were incorrect for the command")
	errRuntimeError = errors.New("There was a problem executing the command")
}
