package utils

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/AbGuthrie/goquery/models"
)

// ScheduleQueryAndWait schedules the provided query with the proved API, and implements blocking
// with a ctrl C interupt
func ScheduleQueryAndWait(api models.GoQueryAPI, uuid, query string) (models.Rows, error) {
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
