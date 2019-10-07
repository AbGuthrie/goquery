package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/utils"
)

// TODO .query should map to Query which is blocking

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
