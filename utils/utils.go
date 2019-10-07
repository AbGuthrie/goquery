// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"github.com/AbGuthrie/goquery/config"
)

// PrettyPrintQueryResults prints a given []result map set to standard out
// taking into consideration the current state.go's print mode
func PrettyPrintQueryResults(results []map[string]string) {
	currentConfig := config.GetConfig()
	switch currentConfig.CurrentPrintMode {
	case "json":
		prettyPrintQueryResultsJSON(results)
	case "line":
		prettyPrintQueryResultsLines(results)
	case "csv":
		prettyPrintQueryResultsCSV(results)
	default:
	}
}
