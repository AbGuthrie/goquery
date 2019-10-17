package commands

import (
	"fmt"
	"strings"

	"github.com/AbGuthrie/goquery/hosts"
)

func printHosts(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This commands takes no parameters")
	}

	hosts := hosts.GetCurrentHosts()
	paddings := calculatePaddings(hosts)

	dividerPadding := 0
	for _, padding := range paddings {
		dividerPadding += padding
	}
	divider := strings.Repeat("-", dividerPadding+15)

	// Print header
	fmt.Printf(
		"%-*s | %-*s | %-*s | %-*s | %-*s | %-*s\n%s\n",
		paddings[0], "UUID",
		paddings[1], "Name",
		paddings[2], "Platform",
		paddings[3], "Version",
		paddings[4], "History count",
		paddings[5], "Username",
		divider,
	)
	// Print rows
	for _, host := range hosts {
		fmt.Printf(
			"%-*s | %-*s | %-*s | %-*s | %*d | %-*s\n",
			paddings[0], host.UUID,
			paddings[1], host.ComputerName,
			paddings[2], host.Platform,
			paddings[3], host.Version,
			paddings[4], len(host.QueryHistory),
			paddings[5], host.Username,
		)
	}
	return nil
}

func calculatePaddings(hosts []hosts.Host) []int {
	maxUUID := 36
	maxHistoryCount := len("History count")

	maxName := len("Name")
	for _, host := range hosts {
		if len(host.UUID) > maxName {
			maxName = len(host.UUID)
		}
	}
	maxPlatform := len("Platform")
	for _, host := range hosts {
		if len(host.ComputerName) > maxPlatform {
			maxPlatform = len(host.UUID)
		}
	}
	maxVersion := len("Version")
	for _, host := range hosts {
		if len(host.Version) > maxVersion {
			maxVersion = len(host.UUID)
		}
	}
	maxUsername := len("Username")
	for _, host := range hosts {
		if len(host.Username) > maxUsername {
			maxUsername = len(host.UUID)
		}
	}

	return []int{
		maxUUID,
		maxName,
		maxPlatform,
		maxVersion,
		maxHistoryCount,
		maxUsername,
	}
}
