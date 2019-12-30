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
	"github.com/AbGuthrie/goquery/commands"
	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/models"

	prompt "github.com/c-bata/go-prompt"
)

func parseConfigOverride(args []string) (string, error) {
	if len(args) == 1 {
		return "", fmt.Errorf("No override provided")
	}

	if len(args) < 3 {
		panic("Invalid arguments provided, expecting --config 'path'")
	}

	if args[1] != "--config" {
		panic("Invalid arguments provided, expecting --config 'path'")
	}

	return args[2], nil
}

func findUserConfig() string {
	configPath, err := parseConfigOverride(os.Args)

	// No config file override provided, check for default in ~/goquery/config.json
	if err != nil {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("Failed to fetch user info for home directory: %s\n", err)
		} else {
			configPath = path.Join(usr.HomeDir, ".goquery/config.json")
		}
	}

	// There is no home folder config so default to system wide
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "/var/goquery/config.json"
	}
	return configPath
}

func loadUserConfig() (config.Config, error) {
	configPath := findUserConfig()
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

func externalExample(api models.GoQueryAPI, config *config.Config, cmdline string) error {
	fmt.Println("Greetings from an external command!")
	return nil
}

func externalExampleHelp() string {
	return "Example external command from outside goquery"
}

func externalExampleSuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}

func main() {
	// 1. Provide something that implements the required models/GoQueryAPI interface,
	//	  or use a supported built in (see `api/mock` for example implementation)
	// api := myCustomAPI{}
	// api, err := osctrl.CreateOSctrlAPI(true)	// import goquery/api/mock
	api, err := mock.CreateMockAPI(true) // import goquery/api/osctrl

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
	commandMap := map[string]commands.GoQueryCommand{
		".external": commands.GoQueryCommand{externalExample, externalExampleHelp, externalExampleSuggest},
		// Possible command that could be used to pull a file from a machine
		//".get": commands.GoQueryCommand{get, getHelp, getSuggest},
	}
	// 3. Call goquery
	goquery.RunWithExternalCommands(api, cfg, commandMap)
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
