// Package utils holds utility functions like print formatting or conversion functions
package utils

import (
	"fmt"
	"strings"
)

// InterpolateArguments fills in an alias' placeholders ($#) with provided arguments
// TODO add alias_test.go unit tests
func InterpolateArguments(rawLine string, command string) (string, error) {
	inputParts := strings.Split(rawLine, " ")
	args := inputParts[1:]

	// TODO this should support escaping and ignoring the
	// placeholder pattern ie \$#
	placeholderParts := strings.Split(command, "$#")

	// Assert arguments provided and placeholders align
	if len(args) != len(placeholderParts)-1 {
		return "", fmt.Errorf("Argument mismatch, alias expects %d args", len(placeholderParts)-1)
	}

	// If no placeholders in query, return as is
	if len(placeholderParts)-1 == 0 {
		return command, nil
	}

	realizedCommand := ""
	for i, arg := range args {
		realizedCommand += placeholderParts[i]
		realizedCommand += arg
	}

	return realizedCommand + placeholderParts[len(placeholderParts)-1], nil
}
