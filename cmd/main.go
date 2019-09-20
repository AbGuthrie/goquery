package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type GoQueryCommand func(string) int

func connect(cmdline string) int {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		fmt.Printf("Host UUID required\n")
		return 1
	}
	fmt.Printf("Connecting to '%s'\n", args[1])
	return 0
}

func exit(cmdline string) int {
	fmt.Printf("Goodbye!\n")
	os.Exit(0)
	return 0
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
			retcode := command(input)
			if retcode != 0 {
				fmt.Printf("%s threw an error!\n", args[0])
			}
		}
	}
}
