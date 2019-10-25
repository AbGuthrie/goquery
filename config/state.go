// Package config is repsonsible for setting and returning the current
// state of the shell in regards to configuration flags and mode options
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
)

// Alias is the struct used to allow abstracted commands
type Alias struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Config is the struct containing the application state
type Config struct {
	CurrentPrintMode PrintMode `json:"printMode"`
	DebugEnabled     bool      `json:"debugEnabled"`
	Aliases          []Alias   `json:"aliases"`
	Experimental     bool      `json:"experimental"`
}

// PrintMode is a type to ensure SetPrintMode recieves a valid enum
type PrintMode int

// PrintMode constants enum
const (
	PrintJSON   PrintMode = 0
	PrintLine   PrintMode = 1
	PrintPretty PrintMode = 2
)

var config Config

func init() {
	configPath, err := parseConfigPath(os.Args)
	if err != nil {
		panic(err)
	}

	// No config file override provided, check for defaul in ~/goquery/config.json
	if configPath == "" {
		usr, err := user.Current()
		if err != nil {
			panic(fmt.Sprintf("Failed to fetch user info for home directory: %s", err))
		}
		goQueryPath := path.Join(usr.HomeDir, ".goquery")
		configPath = path.Join(goQueryPath, "config.json")
	}

	// If no config file exists, use defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		SetPrintMode(PrintPretty)
		return
	}

	// Otherwise go ahead and load + decode config json
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("Unable to read config file: %s", err))
	}
	decoded := &Config{}
	if err := json.Unmarshal(configBytes, &decoded); err != nil {
		panic(fmt.Sprintf("Unable to parse config file: %s", err))
	}

	config = *decoded
	if config.DebugEnabled == true {
		fmt.Printf("Debug mode on\n")
	}

	// Verify loaded aliases
	for _, alias := range config.Aliases {
		if len(strings.Split(name, " ")) > 1 {
			return fmt.Errorf("Aliases must not contain spaces")
		}
		if AliasIsCyclic(alias) {
			// TODO
		}
	}
	// TODO on alias load and parse, assert that no cyclical aliases
	// TODO ensure aliases have no spaces in the command name

}

// GetConfig returns a copy of the current state struct
func GetConfig() Config {
	return config
}

// SetDebug assigns .Debug on the current config struct
func SetDebug(enabled bool) {
	config.DebugEnabled = enabled
}

func GetDebug() bool {
	return config.Debug
}

func SetExperimental(enabled bool) {
	config.Experimental = enabled
}

func GetExperimental() bool {
	return config.Experimental
}

// SetPrintMode assigns .CurrentPrintMode on the current config struct
func SetPrintMode(printMode PrintMode) {
	config.CurrentPrintMode = printMode
}

func parseConfigPath(args []string) (string, error) {
	if len(args) == 1 {
		return "", nil
	}
	// Drop leading `main.go`
	args = args[1:]

	// Currently this is the only command line argument, so it
	// doesn't need to be as robust
	if len(args) < 2 {
		return "", fmt.Errorf("Invalid arguments provided, expecting --config 'path'")
	}
	if args[0] != "--config" {
		return "", fmt.Errorf("Invalid arguments provided, expecting --config 'path'")
	}
	argPath := args[1]
	if _, err := os.Stat(argPath); os.IsNotExist(err) {
		return "", fmt.Errorf("File '%s' does not exist", argPath)
	}
	return argPath, nil
}

// AddAlias adds registers a new alias in the config
func AddAlias(name, command string) error {
	if len(strings.Split(name, " ")) > 1 {
		return fmt.Errorf("Aliases must not contain spaces")
	}
	newAlias := Alias{
		Name:    name,
		Command: command,
	}
	// Check is cyclic
	if AliasIsCyclic(newAlias) {
		return fmt.Errorf("Alias creates an infinite loop")
	}
	config.Aliases = append(config.Aliases, newAlias)
	return nil
}

// AliasIsCyclic is a check to ensure that the aliases do not form an infinite loop
func AliasIsCyclic(alias Alias) bool {
	// TODO
	return false
}
