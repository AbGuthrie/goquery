package main

import (
	"bufio"
	"fmt"
	"github.com/AbGuthrie/goquery/commands"
	"os"
	"strings"
)

func main() {
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
		if command, ok := commands.CommandMap[args[0]]; ok {
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
