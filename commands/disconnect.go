package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/hosts"

	prompt "github.com/c-bata/go-prompt"
)

func disconnect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Host UUID required")
	}
	uuid := args[1]

	if err := hosts.Disconnect(uuid); err != nil {
		return fmt.Errorf("Error disconnecting from host: %s", err)
	}
	fmt.Printf("Disconnected from '%s'\n", uuid)

	return nil
}

func disconnectHelp() string {
	return "Disconnect from a host with UUID"
}

func disconnectSuggest(cmdline string) []prompt.Suggest {
	prompts := []prompt.Suggest{}
	for _, host := range hosts.GetCurrentHosts() {
		prompts = append(prompts, prompt.Suggest{
			Text:        host.UUID,
			Description: host.ComputerName,
		})
	}
	return prompts
}
