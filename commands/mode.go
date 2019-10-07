package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/config"
)

func changeMode(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Mode parameter required")
	}
	modeArg := args[1]

	// Assert valid mode
	valid := false
	for _, mode := range config.PrintModes {
		if mode == modeArg {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("%s is not a valid print mode", modeArg)
	}

	config.SetPrintMode(modeArg)
	fmt.Printf("Print mode set to '%s'.\n", modeArg)

	return nil
}
