// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"encoding/json"
	"fmt"
	"sort"
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
		sortedKeys := sortedColumnKeys(row)
		for _, key := range sortedKeys {
			fmt.Printf("%*s = %s\n", keyPadding, key, row[key])
		}
		fmt.Printf("\n")
	}
}

func sortedColumnKeys(results map[string]string) []string {
	keys := make([]string, 0)
	for key := range results {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
