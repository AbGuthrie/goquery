package models

import (
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"
)

// GoQueryAPI defines the set of functions needed for goquery to interface with a backend
// These functions also must handle any needed authentication because the rest of goquery
// is blind to the implementation for code separation purposes.
type GoQueryAPI interface {
	CheckHost(string) (hosts.Host, error)
	ListHosts() (utils.Rows, error)
	ScheduleQuery(string, string) (string, error)
	FetchResults(string) (utils.Rows, string, error)
}
