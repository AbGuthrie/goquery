package commands

import (
	"fmt"
	"os"

	prompt "github.com/c-bata/go-prompt"
)

func exit(cmdline string) error {
	fmt.Printf("Goodbye!\n")
	os.Exit(0)
	return errRuntimeError
}

func exitHelp() string {
	return "Exit goquery"
}

func exitSuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
