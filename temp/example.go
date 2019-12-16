package main

import (
	"fmt"
	"net/url"

	"github.com/AbGuthrie/goquery"
	"github.com/AbGuthrie/goquery/api/mock"
	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/models"
)

func main() {
	// 1. Provide something that implements the required models/GoQueryAPI interface,
	//	  or use a supported built in (see `api/mock` for example implementation)
	// api := myCustomAPI{}
	// api, err := osctrl.CreateOSctrlAPI(true) // import goquery/api/mock
	api, err := mock.CreateMockAPI(true) // import goquery/api/osctrl
	if err != nil {
		panic(err)
	}

	// 2. Create goquery configuration options (aliases, print mode, debug etc.)
	config := config.Config{
		PrintMode:    "pretty",
		DebugEnabled: true,
		Aliases: map[string]config.Alias{
			".all": config.Alias{
				Description: "Select everything from a table",
				Command:     ".query select * from $#",
			},
		},
	}

	// 3. Call goquery
	goquery.Run(api, config)
}

type myCustomAPI struct {
	url url.URL
}

// Implement GoQueryAPI interface
func (apiConfig myCustomAPI) CheckHost(uuid string) (hosts.Host, error) {
	return hosts.Host{}, fmt.Errorf("Not implemented")
}

func (apiConfig myCustomAPI) ScheduleQuery(uuid string, query string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

func (apiConfig myCustomAPI) FetchResults(queryToken string) (models.Rows, string, error) {
	return models.Rows{}, "", fmt.Errorf("Not implemented")
}
