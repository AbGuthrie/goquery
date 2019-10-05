package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/AbGuthrie/goquery/api"
	"github.com/AbGuthrie/goquery/hosts"
)

var verificationTemplate = "select * from file where path = '%s' and type = 'directory'"

// TODO cd needs to support quoting so you can CD into folders with spaces

func changeDirectory(cmdline string) error {
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return fmt.Errorf("No host is currently connected: %s", err)
	}

	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Directory must be provided")
	}

	// TODO This needs to support Unicode/Runes
	// TODO Support fast mode that doesn't do directory verification
	requestedDirectory := cmdline[strings.Index(cmdline, " ")+1:]

	if len(requestedDirectory) == 0 {
		return fmt.Errorf("Directory requested is invalid")
	}

	// The change isn't absolute so we need the current directory
	if requestedDirectory[0] != '/' {
		requestedDirectory = host.CurrentDirectory + requestedDirectory
	}

	// All directory changes must end with a forward slash
	if requestedDirectory[len(requestedDirectory)-1] != '/' {
		requestedDirectory += "/"
	}

	verificationQuery := fmt.Sprintf(verificationTemplate, requestedDirectory)
	queryName, err := api.ScheduleQuery(host.UUID, verificationQuery)

	if err != nil {
		return err
	}

	// TODO This should be debugging output
	//fmt.Printf("Query Started With Name: %s\n", queryName)

	// Wait while the query is pending
	var results []map[string]string
	var status string
	for {
		results, status, err = api.FetchResults(queryName)
		if err != nil || status != "Pending" {
			break
		}
		time.Sleep(2 * time.Second)
		fmt.Printf(".")
	}

	fmt.Printf("\n")
	if err != nil {
		return err
	}

	if len(results) == 1 {
		return hosts.SetCurrentHostDirectory(requestedDirectory)
	} else {
		return fmt.Errorf("No such directory")
	}

	return nil
}
