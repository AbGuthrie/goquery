package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/models"
	"github.com/AbGuthrie/goquery/utils"

	prompt "github.com/c-bata/go-prompt"
)

func printHosts(api models.GoQueryAPI, config *config.Config, cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This command takes no parameters")
	}

	hostRows := make([]map[string]string, 0)

	for _, host := range hosts.GetCurrentHosts() {
		hostRows = append(hostRows, map[string]string{
			"UUID":              host.UUID,
			"Name":              host.ComputerName,
			"Platform":          host.Platform,
			"Version":           host.Version,
			"Current Directory": host.CurrentDirectory,
			"Username":          host.Username,
		})
	}

	utils.PrettyPrintQueryResults(hostRows, config.PrintMode)

	return nil
}

func printHostsHelp() string {
	return "Prints out all connected hosts"
}

func printHostsSuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
