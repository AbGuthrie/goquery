package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/utils"

	prompt "github.com/c-bata/go-prompt"
)

func discover(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) != 1 {
		return fmt.Errorf("This command takes no parameters")
	}

	// Query for list of available hosts
	enrolledHosts, err := api.ListHosts()
	if err != nil {
		return fmt.Errorf("Error querying available hosts: %s", err)
	}

	utils.PrettyPrintQueryResults(enrolledHosts)

	return nil
}

func discoverHelp() string {
	return "Prints all hosts registered with api server"
}

func discoverSuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
