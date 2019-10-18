package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"
)

func printHosts(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This command takes no parameters")
	}

	hostRows := make([]map[string]string, 0)

	for _, host := range hosts.GetCurrentHosts() {
		hostRows = append(hostRows, map[string]string{
			"UUID" : host.UUID,
			"Name" : host.ComputerName,
			"Platform" : host.Platform,
			"Version" : host.Version,
			"Current Directory" : host.CurrentDirectory,
			"Username" : host.Username,
		})
	}

	utils.PrettyPrintQueryResults(hostRows)

	return nil
}
