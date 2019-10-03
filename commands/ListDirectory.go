package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"
)

// TODO .query should map to Query which is blocking

func ListDirectory(cmdline string) error {
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return fmt.Errorf("No host is currently connected: %s", err)
	}

	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) != 1 {
		return fmt.Errorf("This commands takes no parameters")
	}
	listQuery := fmt.Sprintf("select * from file where directory = '%s'", host.CurrentDirectory)
	queryName, err := api.ScheduleQuery(host.UUID, listQuery)

	if err != nil {
		return err
	}

	// TODO This should be debugging output
	// fmt.Printf("Query Started With Name: %s\n", queryName)
	for {
		_, status, err := api.FetchResults(queryName)
		if err != nil || status != "Pending" {
			break
		}
		time.Sleep(2 * time.Second)
		fmt.Printf(".")
	}
	fmt.Printf("\n")
	results, _, err := api.FetchResults(queryName)
	if err != nil {
		return err
	}

	utils.PrettyPrintQueryResults(results, 0)

	return nil
}
