// Package config is repsonsible for setting and returning the current
// state of the shell in regards to configuration flags and mode options
package config

// Config is the struct returned
type Config struct {
	CurrentPrintMode string
	Debug            bool
}

// PrintMode is a type to ensure SetPrintMode recieves a valid enum
type PrintMode string

// PrintModes
var PrintModes = []string {
	"json",
	"line",
	"csv",
}

var config Config

func init() {
	// TODO this module should be able to load config
	// defaults from a .config file in ~/.goquery
	// and should configure host aliases or default hosts
}

// GetConfig returns a copy of the current state struct
func GetConfig() Config {
	return config
}

// SetDebug assigns .Debug on the current config struct
func SetDebug(enabled bool) {
	config.Debug = enabled
}

// SetPrintMode assigns .CurrentPrintMode on the current config struct
func SetPrintMode(printMode string) {
	config.CurrentPrintMode = printMode
}
