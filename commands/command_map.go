package commands

import (
	"errors"
)

type GoQueryCommand func(string) error

// Function Map
var CommandMap map[string]GoQueryCommand

// Errors
var errArgumentError error
var errRuntimeError error

func init(){
	CommandMap = map[string]GoQueryCommand{
		".connect": connect,
		".exit": exit,
	}

	errArgumentError = errors.New("The arguments provided were incorrect for the command");
	errRuntimeError = errors.New("There was a problem executing the command");
}
