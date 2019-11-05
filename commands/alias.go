package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/config"
	prompt "github.com/c-bata/go-prompt"
)

func alias(cmdline string) error {
	args := strings.Split(cmdline, " ")

	// If no args provided, print current state of aliases
	if len(args) == 1 {
		aliases := config.GetConfig().Aliases
		if len(aliases) == 0 {
			fmt.Printf("No aliases set\n")
			return nil
		}
		fmt.Printf("Available aliases:\n\n")
		for _, alias := range aliases {
			fmt.Printf("Name:    %s\nCommand: %s\n\n", alias.Name, alias.Command)
		}
		return nil
	}

	// If '--add' argument provided, try remove alias from config
	if args[1] == "--add" {
		if len(args) < 4 {
			return fmt.Errorf("--add flag requires an alias arguments: ALIAS_NAME ALIAS_COMMAND")
		}
		args = args[2:]
		name := args[0]
		command := ""
		if len(args) > 1 {
			command = strings.Join(args[1:], " ")
		}

		// Create the command and store in state
		err := config.AddAlias(name, command)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error creating alias: %s\n", err))
		}

		fmt.Printf("Created new alias '%s' with command: %s\n", name, command)
		return nil
	}

	// Handled --add by now, is input any other than '--remove' ?
	if args[1] != "--remove" {
		return fmt.Errorf(".alias must be called with either '--add' or '--remove' flags")
	}
	if len(args) == 2 {
		return fmt.Errorf("--remove flag requires an alias name argument")
	}

	// Argument provided, try remove alias from config
	if err := config.RemoveAlias(args[2]); err != nil {
		return err
	}
	fmt.Printf("Successfully removed alias\n")
	return nil
}

func aliasHelp() string {
	return "Create a new alias or call with no arguments to list current aliases. " +
		"The format for creating an alias is as follows: ALIAS_NAME .example arg1 $# arg3" +
		"To remove an alias, use .alias --remove ALIAS_NAME"
}

func aliasSuggest(cmdline string) []prompt.Suggest {
	// If just at .alias, suggest the flags
	args := strings.Split(cmdline, " ")
	if len(args) == 2 && args[1] == "" {
		return []prompt.Suggest{
			prompt.Suggest{Text: "--add", Description: "Use this flag to create a new alias"},
			prompt.Suggest{Text: "--remove", Description: "Use this flag to remove an alias by name"},
		}
	}
	return []prompt.Suggest{}
}
