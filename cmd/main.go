package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Errors
var errArgumentError = errors.New("The arguments provided were incorrect for the command");
var errRuntimeError = errors.New("There was a problem executing the command");

type GoQueryCommand func(string) error

func connect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		fmt.Errorf("Host UUID required\n")
		return errArgumentError
	}
	fmt.Printf("Connecting to '%s'\n", args[1])
	return nil
}

func exit(cmdline string) error {
	fmt.Printf("Goodbye!\n")
	os.Exit(0)
	return errRuntimeError
}

func main() {
	// Function Map
	commandMap := make(map[string]GoQueryCommand)
	commandMap[".connect"] = connect
	commandMap[".exit"] = exit

	reader := bufio.NewReader(os.Stdin)
	for {
		// Read the keyboad input.
		fmt.Print("goquery> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		input = strings.TrimSuffix(input, "\n") // Remove the newline character
		args := strings.Split(input, " ")       // Separate command and arguments
		if command, ok := commandMap[args[0]]; ok {
			err := command(input)
			if err != nil {
				fmt.Printf("%s: %s!\n", args[0], err.Error())
			}
		} else {
			fmt.Printf("No such command: %s\n", args[0])
			continue
		}
	}
}
