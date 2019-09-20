package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("> ")
		
		// Read the keyboad input.
        input, err := reader.ReadString('\n')
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }

		input = strings.TrimSuffix(input, "\n")		// Remove the newline character		
		args := strings.Split(input, " ")			// Separate command and arguments

		switch args[0] {
		case "connect":
			if len(args) == 1 {
				fmt.Printf("Target address required\n")
				continue
			}
			fmt.Printf("Connecting to '%s'\n", args[1])
		case "exit":
			os.Exit(0)
		default:
			// Echo
			fmt.Printf("Echo %s\n", args[0])
		}
    }
}
