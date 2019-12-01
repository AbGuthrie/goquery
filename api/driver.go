// Package api defines the api interface functions required to implement goquery calls
package api

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/AbGuthrie/goquery/api/mock"
	"github.com/AbGuthrie/goquery/api/models"
	"github.com/AbGuthrie/goquery/api/osctrl"
	"github.com/AbGuthrie/goquery/utils"

	"github.com/AbGuthrie/goquery/hosts"
)

var api models.GoQueryAPI

// InitializeAPI initializes and holds on to the specified instance/implementation
// of the requiredmodels.GoQueryAPI interface
func InitializeAPI(apiName string) error {
	switch apiName {
	case "mock":
		instance, err := mock.Intialize()
		if err != nil {
			return fmt.Errorf("Error initializing mock API %s", err)
		}
		api = instance
		break
	case "osctrl":
		instance, err := osctrl.Initialize()
		if err != nil {
			return fmt.Errorf("Error initializing osctrl API %s", err)
		}
		api = instance
		break
	default:
		return fmt.Errorf("Unknown API implementation: %s", apiName)
	}
	return nil
}

// CheckHost queries the api to validate the UUID references a valid, active node
func CheckHost(uuid string) (hosts.Host, error) {
	return api.CheckHost(uuid)
}

// ListHosts asks the server for all the enrolled osquery nodes
func ListHosts() (utils.Rows, error) {
	return api.ListHosts()
}

// ScheduleQuery posts a query for the target host that osquery will poll for
func ScheduleQuery(uuid string, query string) (string, error) {
	return api.ScheduleQuery(uuid, query)
}

// FetchResults checks the api for the status and results body of a query
func FetchResults(query string) (utils.Rows, string, error) {
	return api.FetchResults(query)
}

// ScheduleQueryAndWait implements ctrl C interupt for required blocking api calls
func ScheduleQueryAndWait(uuid, query string) (utils.Rows, error) {
	ctrlcChannel := make(chan os.Signal, 1)
	signal.Notify(ctrlcChannel, os.Interrupt)
	results := make([]map[string]string, 0)
	queryName, err := api.ScheduleQuery(uuid, query)
	if err != nil {
		return results, fmt.Errorf("ScheduleQueryAndWait call failed: %s", err)
	}

	// Wait while the query is pending
	var status string
	for {
		results, status, err = api.FetchResults(queryName)
		if err != nil || status != "Pending" {
			break
		}
		time.Sleep(time.Second)
		fmt.Printf(".")
		select {
		case <-ctrlcChannel:
			return results, fmt.Errorf("Waiting Cancelled")
		default:
		}
	}

	fmt.Printf("\n")
	// No need to check error here because return is the same
	return results, err
}
