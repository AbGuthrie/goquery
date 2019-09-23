package commands

import (
	"fmt"
	"strings"
	"github.com/AbGuthrie/goquery/api"
)

func connect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Host UUID required\n")
	}
	fmt.Printf("Connecting to '%s'\n", args[1])

	err := api.CheckHost(args[1])

	if err != nil {
		return errRuntimeError
	}

	return nil
}
