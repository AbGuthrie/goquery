// Package hosts is responsible for holding the state of which hosts
// the goquery shell is currently connected to. The state should
// only be mutated via the .connect or .switch commands, but can be
// looked up anywhere.
package hosts

import (
	"fmt"
)

type Host struct {
	UUID             string
	HostName         string
	QueryHistory     []string
	CurrentDirectory string
	Username         string
}

var currentHostIndex int
var connectedHosts []Host

func init() {
	currentHostIndex = -1
	connectedHosts = []Host{}
}

// Register is responsible for adding a host to the list
// of established connected hosts in the host list. Also
// update the cursor of the current connected host.
// If a given host is already in the list, return the index
func Register(uuid string) error {
	for i, host := range connectedHosts {
		if uuid == host.UUID {
			currentHostIndex = i
			return nil
		}
	}
	connectedHosts = append(connectedHosts, Host{
		UUID: uuid,
	})
	currentHostIndex = len(connectedHosts) - 1
	return nil
}

// Disconnect is responsible for removing a host from the list
// Can be called with a specific host uuid or an empty "" to
// denote the current host the cursor is on
func Disconnect(uuid string) error {
	index := -1
	if uuid == "" {
		index = currentHostIndex
	} else {
		// find provided uuid index in list
		for i, host := range connectedHosts {
			if uuid == host.UUID {
				index = i
				break
			}
		}
	}
	if index == -1 {
		return fmt.Errorf("No active host connection with uuid %s", uuid)
	}
	// Remove found host index from list of connected hosts
	connectedHosts = append(connectedHosts[:index], connectedHosts[index+1:]...)
	currentHostIndex = -1
	return nil
}

// SetCurrentHost updates the current index used to fetch
// the uuid of GetCurrentHost's call, returns the uuid
func SetCurrentHost(targetIndex int) (string, error) {
	if targetIndex >= 0 && targetIndex < len(connectedHosts) {
		currentHostIndex = targetIndex
		return connectedHosts[currentHostIndex].UUID, nil
	}
	return "", fmt.Errorf("Index out of range, currently connected to %d host(s)", len(connectedHosts))
}

// GetCurrentHost is a public API that returns the uuid of
// the shell's current host state which is ultimately used
// to prepend a given API call
func GetCurrentHost() (string, error) {
	if len(connectedHosts) == 0 {
		return "", fmt.Errorf("No active host connections")
	}
	if currentHostIndex == -1 {
		return "", fmt.Errorf("No host index set")
	}
	return connectedHosts[currentHostIndex].UUID, nil
}
