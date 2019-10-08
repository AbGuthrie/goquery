package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/config"
)

var validModes = map[string]config.PrintMode{
	"json":   config.PrintJSON,
	"line":   config.PrintLine,
	"pretty": config.PrintPretty,
}

func changeMode(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Mode parameter required")
	}
	modeArg := args[1]

	// Assert valid mode
	mode, ok := validModes[modeArg]
	if !ok {
		return fmt.Errorf("%s is not a valid print mode", modeArg)
	}

	config.SetPrintMode(mode)
	fmt.Printf("Print mode set to '%s'.\n", modeArg)

	return nil
}
