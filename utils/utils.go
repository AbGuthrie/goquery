// Holds utility functions like print formatting or conversion functions
package utils

import (
	"encoding/json"
	"fmt"
)

func prettyPrintQueryResultsJSON(results []map[string]string) {
	formatted, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		fmt.Printf("Could not format query results.\n")
		return
	}
	fmt.Printf("%s\n", formatted)
}

func PrettyPrintQueryResults(results []map[string]string, format int) {
	switch format {
	case 0:
		prettyPrintQueryResultsJSON(results)
	default:
	}
}
