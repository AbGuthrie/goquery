package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func init() {}

func CheckHost(uuid string) error {
	var client = &http.Client{Timeout: time.Second * 10}
	response, err := client.PostForm("http://127.0.0.1:8001/CheckHost",
		url.Values{"uuid": {uuid}},
	)
	if err != nil {
		return fmt.Errorf("CheckHost call failed: %s", err)
	}

	defer response.Body.Close()

	ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("CheckHost response invalid: %s", err)
	}
	return nil
}
