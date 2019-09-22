package commands

import (
	"fmt"
	"strings"
	"time"
	"net/http"
	"io/ioutil"
)

func connect(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Host UUID required\n")
	}
	fmt.Printf("Connecting to '%s'\n", args[1])

	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Get("http://127.0.0.1:8001/status")	// TODO parameterize url in config
	if err != nil {
		// TODO check status code
		return fmt.Errorf("Failed to connect %s'", err)
	}

	// Deserialize body bytes
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Failed to parse response body %s'", err)
	}
	fmt.Printf("Got response'%s'\n", string(body))

	return nil
}
