package api

import (
	"fmt"
	"os"
	"time"

	"github.com/AbGuthrie/goquery/api/mock"
	"github.com/AbGuthrie/goquery/api/osctrl"
	"github.com/AbGuthrie/goquery/api/private"

	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
)

var ctrlcChannel (chan os.Signal)

type Rows = []map[string]string

// GoQueryAPI defines the set of functions needed for goquery to interface with a backend
// These functions also must handle any needed authentication because the rest of goquery
// is blind to the implementation for code separation purposes.
type GoQueryAPI struct {
	CheckHost     func(string) (hosts.Host, error)
	ScheduleQuery func(string, string) (string, error)
	FetchResults  func(string) (Rows, string, error)
}

// CommandMap is the mapping from command line string to GoQueryCommand
var ApiMap map[string]GoQueryAPI

func init() {
	ApiMap = map[string]GoQueryAPI{
		"mock":    GoQueryAPI{mock.CheckHost, mock.ScheduleQuery, mock.FetchResults},
		"osctrl":  GoQueryAPI{osctrl.CheckHost, osctrl.ScheduleQuery, osctrl.FetchResults},
		"private": GoQueryAPI{private.CheckHost, private.ScheduleQuery, private.FetchResults},
	}
	ctrlcChannel = make(chan os.Signal, 1)
}

func CheckHost(uuid string) (hosts.Host, error) {
	if _, ok := ApiMap[config.GetApi()]; !ok {
		return hosts.Host{}, fmt.Errorf("Unknown API: %s", config.GetApi())
	}
	return ApiMap[config.GetApi()].CheckHost(uuid)
}

func ScheduleQuery(uuid string, query string) (string, error) {
	if _, ok := ApiMap[config.GetApi()]; !ok {
		return "", fmt.Errorf("Unknown API: %s", config.GetApi())
	}
	return ApiMap[config.GetApi()].ScheduleQuery(uuid, query)
}

func FetchResults(query string) (Rows, string, error) {
	if _, ok := ApiMap[config.GetApi()]; !ok {
		return Rows{}, "", fmt.Errorf("Unknown API: %s", config.GetApi())
	}
	return ApiMap[config.GetApi()].FetchResults(query)
}

func ScheduleQueryAndWait(uuid string, query string) ([]map[string]string, error) {
	queryName, err := ApiMap[config.GetApi()].ScheduleQuery(uuid, query)
	var results = make([]map[string]string, 0)
	if err != nil {
		return results, fmt.Errorf("ScheduleQueryAndWait call failed: %s", err)
	}

	// Wait while the query is pending
	var status string
	for {
		results, status, err = ApiMap[config.GetApi()].FetchResults(queryName)
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
