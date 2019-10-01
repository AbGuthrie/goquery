package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
)

func ScheduleQuery(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("A query to run must be provided.\n")
	}
	// TODO This needs to support Unicode/Runes
	commandStripped := cmdline[strings.Index(cmdline, " ") + 1:]
	fmt.Printf("Running '%s'\n", commandStripped)

	queryName, err := api.ScheduleQuery("", commandStripped)

	if err != nil {
		return err
	}

	fmt.Printf("Query Started With Name: %s\n", queryName)

	return nil
}
