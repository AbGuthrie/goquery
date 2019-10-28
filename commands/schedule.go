package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"

	prompt "github.com/c-bata/go-prompt"
)

func schedule(cmdline string) error {
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
	queryName, err := api.ScheduleQuery(host.UUID, commandStripped)

	if err != nil {
		return err
	}

	fmt.Printf("Scheduled query for host. Resume with name: %s\n", queryName)

	return nil
}

func scheduleHelp() string {
	return "Schedule a query on a host but don't wait for results"
}

func scheduleSuggest(cmdline string) []prompt.Suggest {
	return querySuggest(cmdline)
}
