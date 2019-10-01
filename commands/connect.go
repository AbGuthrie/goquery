package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"
)

func connect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Host UUID required")
	}
	uuid := args[1]
	fmt.Printf("Connecting to '%s'\n", uuid)

	err := api.CheckHost(uuid)

	if err != nil {
		return err
	}

	// All is good, update hosts state
	if err := hosts.Register(uuid); err != nil {
		return fmt.Errorf("Error registering host: %s", err)
	}
	fmt.Println("Successfully connected.")

	return nil
}
