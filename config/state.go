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
	Name    string
	Command string `json:"command"`
}

// Config is the struct containing the application state
type Config struct {
	CurrentPrintMode PrintMode        `json:"printMode"`
	DebugEnabled     bool             `json:"debugEnabled"`
	Aliases          map[string]Alias `json:"aliases"`
	Experimental     bool             `json:"experimental"`
	Api              string           `json:"api"`
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
			fmt.Printf("Failed to fetch user info for home directory: %s\n", err)
		}
		goQueryPath := path.Join(usr.HomeDir, ".goquery")
		configPath = path.Join(goQueryPath, "config.json")
	}

	// If no config file exists, use defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		SetPrintMode(PrintPretty)
		config.Api = "mock"
		return
	}

	// Otherwise go ahead and load + decode config json
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Unable to read config file: %s at path %s\n", err, configPath)
	}
	decoded := &Config{}
	if err := json.Unmarshal(configBytes, &decoded); err != nil {
		fmt.Printf("Unable to parse config file: %s at path %s\n", err, configPath)
	}

	config = *decoded

	// Validate/filter loaded aliases
	validAliases := map[string]Alias{}
	for aliasName, alias := range config.Aliases {
		if len(strings.Fields(aliasName)) > 1 {
			fmt.Printf("Aliases error: Name '%s' must not contain whitespace\n", aliasName)
			continue
		}
		if AliasIsCyclic(alias) {
			fmt.Printf("Alias error: '%s' creates an infinite loop\n", aliasName)
			continue
		}
		validAliases[aliasName] = Alias{
			Name:    aliasName,
			Command: alias.Command,
		}
	}
	config.Aliases = validAliases

	if config.DebugEnabled {
		fmt.Println("Debug mode on")
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
	return config.DebugEnabled
}

func SetExperimental(enabled bool) {
	config.Experimental = enabled
}

func GetExperimental() bool {
	return config.Experimental
}

func GetApi() string {
	return config.Api
}

// SetPrintMode assigns .CurrentPrintMode on the current config struct
func SetPrintMode(printMode PrintMode) {
	config.CurrentPrintMode = printMode
}

// TODO recieve args as map via use a command line parsing library called in from main
func parseConfigPath(args []string) (string, error) {
	if len(args) == 1 {
		return "", nil
	}
	// Drop leading `main.go`
	args = args[1:]

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
	if len(strings.Fields(name)) > 1 {
		return fmt.Errorf("Alias name must not contain any whitespace")
	}
	// Ensure not conflicting names
	if _, exists := config.Aliases[name]; exists {
		return fmt.Errorf("Aliases name '%s' is a duplicate of an existing alias", name)
	}
	newAlias := Alias{
		Name:    name,
		Command: command,
	}
	// Check is cyclic
	if AliasIsCyclic(newAlias) {
		return fmt.Errorf("Alias creates an infinite loop")
	}
	config.Aliases[name] = newAlias
	return nil
}

// RemoveAlias an alias in the config
func RemoveAlias(name string) error {
	if _, exists := config.Aliases[name]; !exists {
		return fmt.Errorf("Alias '%s' not found", name)
	}
	delete(config.Aliases, name)
	return nil
}

// AliasIsCyclic is a check to ensure that the aliases do not form an infinite loop
func AliasIsCyclic(alias Alias) bool {
	graph := make(map[string]string)
	graph[alias.Name] = strings.Split(alias.Command, " ")[0]

	for _, alias := range config.Aliases {
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
