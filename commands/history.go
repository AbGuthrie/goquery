package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/hosts"
)

func history(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This commands takes no parameters")
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
