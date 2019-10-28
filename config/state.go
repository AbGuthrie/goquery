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
type PrintMode string

// PrintMode constants enum
const (
	PrintJSON   PrintMode = "json"
	PrintLine   PrintMode = "line"
	PrintPretty PrintMode = "pretty"
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

	// Validate/filter loaded aliases
	validAliases := []Alias{}
	for _, alias := range config.Aliases {
		if len(strings.Split(alias.Name, " ")) > 1 {
			fmt.Printf("Aliases error: Name '%s' must not contain spaces\n", alias.Name)
			continue
		}
		for _, existing := range validAliases {
			if alias.Name == existing.Name {
				fmt.Printf("Aliases name '%s' is a duplicate of an existing alias\n", alias.Name)
				continue
			}
		}
		if AliasIsCyclic(alias) {
			fmt.Printf("Alias error: '%s' creates an infinite loop\n", alias.Name)
			continue
		}
		validAliases = append(validAliases, alias)
	}
	config.Aliases = validAliases

	if config.DebugEnabled {
		fmt.Printf("Debug mode on\n")
		fmt.Printf("Initialized with print mode '%s'\n", config.CurrentPrintMode)
		fmt.Printf("Loaded %d aliase(s)\n", len(config.Aliases))
		fmt.Println("")
	}
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
	// Ensure not conflicting names
	for _, alias := range config.Aliases {
		if alias.Name == name {
			return fmt.Errorf("Aliases name '%s' is a duplicate of an existing alias", name)
		}
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

// RemoveAlias an alias in the config
func RemoveAlias(name string) error {
	index := -1
	for i, alias := range config.Aliases {
		if name == alias.Name {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("Alias '%s' not found", name)
	}
	config.Aliases = append(config.Aliases[:index], config.Aliases[index+1:]...)
	return nil
}

// AliasIsCyclic is a check to ensure that the aliases do not form an infinite loop
func AliasIsCyclic(alias Alias) bool {
	graph := make(map[string]string)
	aliases := config.Aliases
	aliases = append(config.Aliases, alias)

	for _, alias := range aliases {
		next := strings.Split(alias.Command, " ")[0]
		graph[alias.Name] = next
	}
	return isCyclic(alias.Name, []string{}, graph)
}

func isCyclic(next string, visited []string, nodes map[string]string) bool {
	// Was next node was already visited?
	for _, visitedCommand := range visited {
		// If so, aliases are cyclic, return
		if next == visitedCommand {
			return true
		}
	}
	// If next is not an alias, all good, no alias loop
	newNext, ok := nodes[next]
	if !ok {
		return false
	}
	// Otherwise, continue traversing through aliases
	visited = append(visited, next)
	return isCyclic(newNext, visited, nodes)
}
