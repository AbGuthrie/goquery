package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path"

	"github.com/AbGuthrie/goquery"
	"github.com/AbGuthrie/goquery/api/mock"
	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/models"
)

func parseConfigPath(args []string) (string, error) {
	if len(args) == 1 {
		return "", nil
	}
	// Drop leading `main.go`
	args = args[1:]

	if len(args) < 2 {
		return "", fmt.Errorf("Invalid arguments provided, expecting --config 'path'")
	}
	if args[0] != "--config" {
		return "", fmt.Errorf("Invalid arguments provided, expecting --config 'path'")
	}
	argPath := args[1]
	if _, err := os.Stat(argPath); os.IsNotExist(err) {
		return "", fmt.Errorf("File '%s' does not exist", argPath)
	}
	return argPath, nil
}

func loadUserConfig() (config.Config, error) {
	configPath, err := parseConfigPath(os.Args)
	if err != nil {
		panic(err)
	}

	// No config file override provided, check for default in ~/goquery/config.json
	if configPath == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("Failed to fetch user info for home directory: %s\n", err)
		}
		goQueryPath := path.Join(usr.HomeDir, ".goquery")
		configPath = path.Join(goQueryPath, "config.json")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config.Config{}, fmt.Errorf("No config file found")
	}

	// Otherwise go ahead and load + decode config json
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Unable to read config file: %s at path %s\n", err, configPath)
	}
	decoded := &config.Config{}
	if err := json.Unmarshal(configBytes, &decoded); err != nil {
		fmt.Printf("Unable to parse config file: %s at path %s\n", err, configPath)
	}

	return *decoded, nil
}

func main() {
	// 1. Provide something that implements the required models/GoQueryAPI interface,
	//	  or use a supported built in (see `api/mock` for example implementation)
	// api := myCustomAPI{}
	// api, err := osctrl.CreateOSctrlAPI(true)	// import goquery/api/mock
	api, err := mock.CreateMockAPI(true)			// import goquery/api/osctrl

	if err != nil {
		fmt.Printf("Encountered an error starting API: %s\n", err)
		return
	}

	// 2. Create goquery configuration options (aliases, print mode, debug etc.)
	// You can load from a file or use a hardcoded config (we use a hardcoded config)
	// on error loading from the user's home folder
	cfg, err := loadUserConfig()
	if err != nil {
		fmt.Printf("Couldn't load user config because of error: %s\n", err)
		fmt.Println("Using defaults")

		cfg = config.Config{
			PrintMode:    "pretty",
			DebugEnabled: true,
			Aliases: map[string]config.Alias{
				".all": config.Alias{
					Description: "Select everything from a table",
					Command:     ".query select * from $#",
				},
			},
		}
	}

	// 3. Call goquery
	goquery.Run(api, cfg)
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
