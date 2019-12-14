package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/models"

	prompt "github.com/c-bata/go-prompt"
)

func history(api models.GoQueryAPI, config config.Config, cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This command takes no parameters")
	}

	host, err := hosts.GetCurrentHost()

	if err != nil {
		return err
	}

	fmt.Printf("Query Name : Query\n")
	for _, query := range host.QueryHistory {
		fmt.Printf("%s: %s\n", query.Name, query.SQL)
	}

	return nil
}

func historyHelp() string {
	return "Print the current host's query history from the current session"
}

func historySuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
