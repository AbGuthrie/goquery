package commands

import (
	"fmt"
	"os"

	"github.com/AbGuthrie/goquery/v2/config"
	"github.com/AbGuthrie/goquery/v2/models"
	prompt "github.com/c-bata/go-prompt"
)

func exit(api models.GoQueryAPI, config *config.Config, cmdline string) error {
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
