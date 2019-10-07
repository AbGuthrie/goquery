package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"
)

// TODO .query should map to Query which is blocking

func listDirectory(cmdline string) error {
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return fmt.Errorf("No host is currently connected: %s", err)
	}

	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) != 1 {
		return fmt.Errorf("This commands takes no parameters")
	}
	listQuery := fmt.Sprintf("select * from file where directory = '%s'", host.CurrentDirectory)
	results, err := api.ScheduleQueryAndWait(host.UUID, listQuery)

	if err != nil {
		return err
	}

	utils.PrettyPrintQueryResults(results)
	return nil
}
