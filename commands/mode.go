package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AbGuthrie/goquery/config"

	prompt "github.com/c-bata/go-prompt"
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

func changeModeHelp() string {
	modeNames := make([]string, 0)
	for mode, _ := range validModes {
		modeNames = append(modeNames, mode)
	}
	sort.Strings(modeNames)
	return fmt.Sprintf("Change print mode (%s)", strings.Join(modeNames, ", "))
}

func changeModeSuggest(cmdline string) []prompt.Suggest {
	prompts := []prompt.Suggest{}

	modeNames := make([]string, 0)
	for mode, _ := range validModes {
		modeNames = append(modeNames, mode)
	}
	sort.Strings(modeNames)

	for _, mode := range modeNames {
		prompts = append(prompts, prompt.Suggest{
			Text:        mode,
			Description: "",
		})
	}

	return prompts
}
