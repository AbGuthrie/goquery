// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"github.com/AbGuthrie/goquery/v2/config"
	"github.com/AbGuthrie/goquery/v2/models"
)

// PrettyPrintQueryResults prints a given []result map set to standard out
// taking into consideration the current state.go's print mode
func PrettyPrintQueryResults(results models.Rows, printMode config.PrintModeEnum) {
	switch printMode {
	case config.PrintJSON:
		prettyPrintQueryResultsJSON(results)
	case config.PrintLine:
		prettyPrintQueryResultsLines(results)
	default:
		prettyPrintQueryResultsPretty(results)
	}
}
