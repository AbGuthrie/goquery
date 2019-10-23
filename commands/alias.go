package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/config"
	prompt "github.com/c-bata/go-prompt"
)

func alias(cmdline string) error {
	// TODO parse and append to config.Aliases
	// TODO on alias add, DAG assert no infinite loop
	return nil

	// TODO .alias with no arguments should print all current aliases
}

func aliasHelp() string {
	return "TODO docs"
}

func aliasSuggest(cmdline string) []prompt.Suggest {
	// TODO easily return list of config.Aliases here
	return []prompt.Suggest{}
}

// TODO docs
func FindAlias(command string) (config.Alias, bool) {
	aliases := config.GetConfig().Aliases
	for _, alias := range aliases {
		if command == alias.Name {
			return alias, true
		}
	}
	return config.Alias{}, false
}

// TODO docs
// TODO add alias_test.go unit tests
func InterpolateArguments(rawLine string, command string) (string, error) {
	inputParts := strings.Split(rawLine, " ")
	args := inputParts[1:]

	// TODO this should support escaping and ignoring the
	// placeholder pattern ie \$#
	placeholderParts := strings.Split(command, "$#")

	// Assert arguments provided and placeholders align
	if len(args) != len(placeholderParts)-1 {
		return "", fmt.Errorf("Argument mismatch, alias expects %d args", len(placeholderParts)-1)
	}

	// If no placeholders in query, return as is
	if len(placeholderParts)-1 == 0 {
		return command, nil
	}

	realizedCommand := ""
	for i, arg := range args {
		realizedCommand += placeholderParts[i]
		realizedCommand += arg
	}

	return realizedCommand, nil
}
