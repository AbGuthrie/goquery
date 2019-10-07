// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"encoding/json"
	"fmt"
)

func prettyPrintQueryResultsJSON(results []map[string]string) {
	fmt.Printf("\n")
	formatted, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		fmt.Printf("Could not format query results.\n")
		return
	}
	fmt.Printf("%s\n", formatted)
}

func prettyPrintQueryResultsCSV(results []map[string]string) {
	if len(results) == 0 {
		return
	}

	for key := range results[0] {
		fmt.Printf("%s,", key)
	}
	fmt.Printf("\n")

	for _, row := range results {
		for _, value := range row {
			fmt.Printf("%s,", value)
		}
		fmt.Printf("\n")
	}
}

func prettyPrintQueryResultsLines(results []map[string]string) {
	fmt.Printf("\n")
	if len(results) == 0 {
		return
	}
	// To center align keys with "=" get longest length key name
	keyPadding := 0
	for key := range results[0] {
		if len(key) > keyPadding {
			keyPadding = len(key)
		}
	}
	for _, row := range results {
		for key, val := range row {
			fmt.Printf("%*s = %s\n", keyPadding, key, val)
		}
		fmt.Printf("\n")
	}
}
