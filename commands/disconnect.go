package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/hosts"
)

func disconnect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Host UUID required")
	}
	uuid := args[1]

	if err := hosts.Disconnect(uuid); err != nil {
		return fmt.Errorf("Error disconnecting from host: %s", err)
	}
	fmt.Printf("Disconnected from '%s'\n", uuid)

	return nil
}
