package types

import (
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/utils"
)

type API interface {
	CheckHost(string) (hosts.Host, error)
	ScheduleQuery(string, string) (string, error)
	FetchResults(string) (utils.Rows, string, error)
	Close() error
}

type History interface {
	GetRecent(int) []string
	Append(string) error
	Close() error
}
