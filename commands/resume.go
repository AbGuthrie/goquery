package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"

	prompt "github.com/c-bata/go-prompt"
)

func resume(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("A query name to resume must be provided")
	}
	// TODO This needs to support Unicode/Runes
	commandStripped := cmdline[strings.Index(cmdline, " ")+1:]
	results, status, err := api.FetchResults(commandStripped)

	if err != nil {
		return err
	}

	if status == "Pending" {
		return fmt.Errorf("Query does not have results available yet")
	}

	utils.PrettyPrintQueryResults(results)

	return nil
}

func resumeHelp() string {
	return "Try to fetch results for a query but don't block if unavailable"
}

func resumeSuggest(cmdline string) []prompt.Suggest {
	host, err := hosts.GetCurrentHost()
	prompts := []prompt.Suggest{}
	if err != nil {
		return prompts
	}

	for _, query := range host.QueryHistory {
		prompts = append(prompts, prompt.Suggest{query.Name, query.SQL})
	}
	return prompts
}
