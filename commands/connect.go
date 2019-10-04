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
	host, err := api.CheckHost(uuid);
	if err != nil {
		return err
	}

	// All is good, update hosts state
	if err := hosts.Register(host); err != nil {
		return fmt.Errorf("Error connecting to host: %s", err)
	}
	fmt.Printf("Successfully connected to '%s'.\n", uuid)

	return nil
}
