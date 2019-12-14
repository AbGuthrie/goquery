// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"github.com/AbGuthrie/goquery/config"
)

// PrettyPrintQueryResults prints a given []result map set to standard out
// taking into consideration the current state.go's print mode
func PrettyPrintQueryResults(results []map[string]string, printMode config.PrintMode) {
	switch printMode {
	case config.PrintJSON:
		prettyPrintQueryResultsJSON(results)
	case config.PrintLine:
		prettyPrintQueryResultsLines(results)
	case config.PrintPretty:
		prettyPrintQueryResultsPretty(results)
	default:
	}
}
