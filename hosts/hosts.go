// Package hosts is responsible for holding the state of which hosts
// the goquery shell is currently connected to. The state should
// only be mutated via the .connect or .switch commands, but can be
// looked up anywhere.
package hosts

import (
	"fmt"
)

type Query struct {
	Name string
	SQL  string
}

type Host struct {
	UUID             string
	ComputerName     string
	Platform         string
	Version          string
	QueryHistory     []Query
	CurrentDirectory string
	Username         string
	Tables           []string
}

func (host *Host) SetCurrentDirectory(newDirectory string) error {
	if len(newDirectory) == 0 {
		return fmt.Errorf("You cannot set directory to empty string")
	}
	if newDirectory[len(newDirectory)-1] != '/' {
		return fmt.Errorf("Final character of directory must be /")
	}
	host.CurrentDirectory = newDirectory
	return nil
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
func Register(newHost Host) error {
	for i, host := range connectedHosts {
		if newHost.UUID == host.UUID {
			currentHostIndex = i
			return nil
		}
	}
	connectedHosts = append(connectedHosts, newHost)
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

// GetCurrentHost is a public API that returns a point to the current host structure.
func GetCurrentHost() (Host, error) {
	if len(connectedHosts) == 0 {
		return Host{}, fmt.Errorf("No active host connections")
	}

	if currentHostIndex == -1 {
		return Host{}, fmt.Errorf("No host index set")
	}
	return connectedHosts[currentHostIndex], nil
}

func SetCurrentHostDirectory(newDirectory string) error {
	if len(connectedHosts) == 0 {
		return fmt.Errorf("No active host connections")
	}

	if currentHostIndex == -1 {
		return fmt.Errorf("No host index set")
	}
	return connectedHosts[currentHostIndex].SetCurrentDirectory(newDirectory)
}

func SetHostTables(uuid string, tables []string) {
	for index, _ := range connectedHosts {
		if connectedHosts[index].UUID != uuid {
			continue
		}
		connectedHosts[index].Tables = tables
		return
	}
	panic("Setting Tables On Unconnected Host!! Something is very wrong!")
}

func AddQueryToHost(uuid string, newQuery Query) {
	for index, _ := range connectedHosts {
		if connectedHosts[index].UUID != uuid {
			continue
		}
		connectedHosts[index].QueryHistory = append(connectedHosts[index].QueryHistory, newQuery)
		return
	}
	panic("Query Ran On Unconnected Host!! Something is very wrong!")
}

// GetCurrentHosts is a public API that returns a the current state of the connectedHosts array
func GetCurrentHosts() []Host {
	return connectedHosts
}
