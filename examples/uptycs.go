package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/AbGuthrie/goquery/v2"
	"github.com/AbGuthrie/goquery/v2/api/uptycs"
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
	if err != nil {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("Failed to fetch user info for home directory: %s\n", err)
		} else {
			configPath = path.Join(usr.HomeDir, ".goquery/config.json")
		}
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "/var/goquery/config.json"
	}
	return configPath
}

func loadUserConfig() (uptycs.GoqueryConfig, error) {
	configPath := findUserConfig()
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Unable to read config file: %s at path %s\n", err, configPath)
	}
	decoded := &uptycs.GoqueryConfig{}
	if err := json.Unmarshal(configBytes, &decoded); err != nil {
		fmt.Printf("Unable to parse config file: %s at path %s\n", err, configPath)
	}
	return *decoded, nil
}

func main() {
	cfg, err := loadUserConfig()
	if err != nil {
		panic(
			fmt.Errorf(
				"Couldn't load user config because of error: %s\n",
				err,
			),
		)
	}
	api, err := uptycs.CreateUptycsAPI(cfg.UptCfgPath, cfg.GoqueryConfig.DebugEnabled)
	if err != nil {
		fmt.Printf("Encountered an error starting API: %s\n", err)
		return
	}
	goquery.Run(api, cfg.GoqueryConfig)
}
