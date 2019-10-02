package api

import (
	"crypto/tls"
	"encoding/json"
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
	response, err := client.PostForm("https://127.0.0.1:8001/checkHost",
		url.Values{"uuid": {uuid}},
	)
	if err != nil {
		return fmt.Errorf("CheckHost call failed: %s", err)
	}
	if response.StatusCode == 200 {
		return nil
	}
	if response.StatusCode == 404 {
		return fmt.Errorf("Unknown Host")
	}
	return fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
}

func ScheduleQuery(uuid string, query string) (string, error) {
	type QueryScheduleResponse struct {
		QueryName string `json:"queryName"`
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr, Timeout: time.Second * 10}
	response, err := client.PostForm("https://127.0.0.1:8001/scheduleQuery",
		url.Values{
			"uuid":  {uuid},
			"query": {query}},
	)
	if err != nil {
		return "", fmt.Errorf("ScheduleQuery call failed: %s", err)
	}
	if response.StatusCode == 404 {
		return "", fmt.Errorf("Unknown Host")
	}
	if response.StatusCode == 200 {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", fmt.Errorf("Could not read response")
		}
		qsResponse := QueryScheduleResponse{}
		err = json.Unmarshal(bodyBytes, &qsResponse)
		if err != nil {
			return "", err
		}
		return qsResponse.QueryName, nil
	}
	return "", fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
}

func FetchResults(queryName string) (string, error) {
	type ResultsResponse struct {
		Rows   []map[string]string `json:"results"`
		Status string              `json:"status"`
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr, Timeout: time.Second * 10}
	response, err := client.PostForm(
		"https://127.0.0.1:8001/fetchResults",
		url.Values{"queryName": {queryName}},
	)
	if err != nil {
		return "", fmt.Errorf("FetchResults call failed: %s", err)
	}
	if response.StatusCode == 404 {
		return "", fmt.Errorf("Unknown queryName")
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Could not read fetchResults response")
	}

	resultsResponse := ResultsResponse{}
	if err := json.Unmarshal(bodyBytes, &resultsResponse); err != nil {
		return "", err
	}
	fmt.Printf("Successfully fetched results:\n%#v\n", resultsResponse)

	// Return QueryResultsResponse type (outer caller should check .Status)
	return "TODO Results coming soon", nil
}
