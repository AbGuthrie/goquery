// Package config is repsonsible for setting and returning the current
// state of the shell in regards to configuration flags and mode options
package config

import (
	"fmt"
	"strings"
)

// Alias is the struct used to allow abstracted commands
type Alias struct {
	Name        string
	Description string `json:"description"`
	Command     string `json:"command"`
}

// Config is the struct containing the application state
type Config struct {
	DebugEnabled bool             `json:"debugEnabled"`
	APIDriver    string           `json:"apiDriver"`
	Experimental bool             `json:"experimental"`
	PrintMode    PrintMode        `json:"printMode"`
	Aliases      map[string]Alias `json:"aliases"`
}

// PrintMode is a type to ensure SetPrintMode recieves a valid enum
type PrintMode string

// PrintMode constants enum
const (
	PrintJSON   PrintMode = "json"
	PrintLine   PrintMode = "line"
	PrintPretty PrintMode = "pretty"
)

// Validate is responsible for filtering incorrect aliases configured, and printing the state of debug modes
func (config *Config) Validate() {
	validAliases := map[string]Alias{}
	for aliasName, alias := range config.Aliases {
		if len(strings.Fields(aliasName)) > 1 {
			fmt.Printf("Aliases error: Name '%s' must not contain whitespace\n", aliasName)
			continue
		}
		if AliasIsCyclic(alias, config.Aliases) {
			fmt.Printf("Alias error: '%s' creates an infinite loop\n", aliasName)
			continue
		}
		validAliases[aliasName] = Alias{
			Name:        aliasName,
			Command:     alias.Command,
			Description: alias.Description,
		}
	}
	config.Aliases = validAliases

	if config.DebugEnabled {
		fmt.Println("Debug mode on")
		fmt.Printf("Initialized with print mode '%s'\n", config.PrintMode)
		fmt.Printf("Loaded %d alias(es)\n", len(config.Aliases))
		fmt.Println("")
	}
}

// SetDebug assigns .Debug on the current config struct
func (config *Config) SetDebug(enabled bool) {
	config.DebugEnabled = enabled
}

func (config *Config) GetDebug() bool {
	return config.DebugEnabled
}

func (config *Config) SetExperimental(enabled bool) {
	config.Experimental = enabled
}

func (config *Config) GetExperimental() bool {
	return config.Experimental
}

func (config *Config) GetAPIDriver() string {
	return config.APIDriver
}

// SetPrintMode assigns .PrintMode on the current config struct
func (config *Config) SetPrintMode(printMode PrintMode) {
	config.PrintMode = printMode
}

// AddAlias adds registers a new alias in the config
func (config *Config) AddAlias(name, command string) error {
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
	if AliasIsCyclic(newAlias, config.Aliases) {
		return fmt.Errorf("Alias creates an infinite loop")
	}
	config.Aliases[name] = newAlias
	return nil
}

// RemoveAlias an alias in the config
func (config *Config) RemoveAlias(name string) error {
	if _, exists := config.Aliases[name]; !exists {
		return fmt.Errorf("Alias '%s' not found", name)
	}
	delete(config.Aliases, name)
	return nil
}

// AliasIsCyclic is a check to ensure that the aliases do not form an infinite loop
func AliasIsCyclic(newAlias Alias, allAliases map[string]Alias) bool {
	graph := make(map[string]string)
	graph[newAlias.Name] = strings.Split(newAlias.Command, " ")[0]

	for _, alias := range allAliases {
		next := strings.Split(alias.Command, " ")[0]
		graph[alias.Name] = next
	}

	return isCyclic(newAlias.Name, []string{}, graph)
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
