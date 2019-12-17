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

func query(api models.GoQueryAPI, config *config.Config, cmdline string) error {
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return fmt.Errorf("No host is currently connected: %s", err)
	}

	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("A query to run must be provided")
	}
	// TODO This needs to support Unicode/Runes
	commandStripped := cmdline[strings.Index(cmdline, " ")+1:]
	results, err := utils.ScheduleQueryAndWait(api, host.UUID, commandStripped)

	if err != nil {
		return err
	}

	utils.PrettyPrintQueryResults(results, config.PrintMode)

	return nil
}

func queryHelp() string {
	return "Schedule a query on a host and wait for results"
}

func querySuggest(cmdline string) []prompt.Suggest {
	parts := strings.Split(cmdline, " ")
	// The cmdline doesn't have enough components
	if len(parts) < 2 {
		return []prompt.Suggest{}
	}

	// If they've anything other than "from "
	if parts[len(parts)-2] != "from" {
		return []prompt.Suggest{}
	}

	prompts := []prompt.Suggest{}

	// There is no connected host
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return prompts
	}
	for _, table := range host.Tables {
		prompts = append(prompts, prompt.Suggest{table, ""})
	}

	return prompts
}
