// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/AbGuthrie/goquery/models"
)

func prettyPrintQueryResultsJSON(results models.Rows) {
	formatted, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		fmt.Printf("Could not format query results.\n")
		return
	}
	fmt.Printf("%s\n", formatted)
}

func prettyPrintQueryResultsLines(results models.Rows) {
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
	sortedKeys := sortedColumnKeys(results[0])
	for _, row := range results {
		for _, key := range sortedKeys {
			fmt.Printf("%*s = %s\n", keyPadding, key, row[key])
		}
		fmt.Printf("\n")
	}
}

func prettyPrintQueryResultsPretty(results models.Rows) {
	maxLens, err := calculateMaxColumnLengths(results)
	if err != nil {
		return
	}

	keyOrder := sortedColumnKeys(results[0])

	dividerLength := 0
	for _, padding := range maxLens {
		dividerLength += padding
	}
	// Then max length of all keys + divider and space (| %s ) + the final divider
	divider := strings.Repeat("-", dividerLength+len(maxLens)*3+1)

	// Print header
	fmt.Printf("%s\n", divider)
	for _, columnName := range keyOrder {
		fmt.Printf("| %-*s ", maxLens[columnName], columnName)
	}
	fmt.Printf("|\n%s\n", divider)

	// Print results
	for _, row := range results {
		for _, columnName := range keyOrder {
			fmt.Printf("| %-*s ", maxLens[columnName], row[columnName])
		}
		fmt.Printf("|\n")
	}

	fmt.Printf("%s\n", divider)
}

func calculateMaxColumnLengths(results models.Rows) (map[string]int, error) {
	if len(results) == 0 {
		return map[string]int{}, fmt.Errorf("Cannot calculate lengths with no rows")
	}

	maxLengths := make(map[string]int)

	// They key may be longer than the values in some cases like inode
	for columnName := range results[0] {
		maxLengths[columnName] = len(columnName)
	}

	for _, row := range results {
		for columnName, rowValue := range row {
			length := len(rowValue)
			if length <= maxLengths[columnName] {
				continue
			}
			maxLengths[columnName] = len(rowValue)
		}
	}

	return maxLengths, nil
}

func sortedColumnKeys(results map[string]string) []string {
	keys := make([]string, 0)
	for key := range results {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
