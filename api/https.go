package api

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func init() {}

func CheckHost(uuid string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr, Timeout: time.Second * 10}
	response, err := client.PostForm("https://127.0.0.1:8001/CheckHost",
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
	if response.StatusCode == 200 {
		return nil
	}
	if response.StatusCode == 404 {
		return fmt.Errorf("Unknown Host")
	}
	return fmt.Errorf("Server returned unknown error: %d", response.StatusCode)

}
