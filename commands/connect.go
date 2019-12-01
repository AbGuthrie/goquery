package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"

	prompt "github.com/c-bata/go-prompt"
)

func connect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Host UUID required")
	}
	uuid := args[1]
	host, err := api.CheckHost(uuid)
	if err != nil {
		return err
	}

	// All is good, update hosts state
	if err := hosts.Register(host); err != nil {
		return fmt.Errorf("Error connecting to host: %s", err)
	}
	fmt.Printf("Verified Host(%s) Exists.\n", uuid)

	results, err := api.ScheduleQueryAndWait(
		host.UUID,
		"select name from osquery_registry where registry = 'table' and active = 1",
	)

	if err != nil {
		return err
	}

	tables := make([]string, 0)
	for _, row := range results {
		// Probably unneeded guard against bad osquery/api data
		if table, ok := row["name"]; ok {
			tables = append(tables, table)
		}
	}
	hosts.SetHostTables(uuid, tables)

	return nil
}

func connectHelp() string {
	return "Connect to a host with UUID"
}

func connectSuggest(cmdline string) []prompt.Suggest {
	prompts := []prompt.Suggest{}
	for _, host := range hosts.GetCurrentHosts() {
		prompts = append(prompts, prompt.Suggest{
			Text:        host.UUID,
			Description: host.ComputerName,
		})
	}
	return prompts
}
