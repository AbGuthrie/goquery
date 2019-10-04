package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/AbGuthrie/goquery/hosts"
)

var ctrlcChannel (chan os.Signal)

func init() {
	ctrlcChannel = make(chan os.Signal, 1)
	signal.Notify(ctrlcChannel, os.Interrupt)
}

func CheckHost(uuid string) (hosts.Host, error) {
	type APIHost struct {
		UUID           string `json:"UUID"`
		ComputerName   string `json:"ComputerName"`
		HostIdentifier string `json:"HostIdentifier"`
		Platform       string `json:"Platform"`
		Version        string `json:"Version"`
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr, Timeout: time.Second * 10}
	response, err := client.PostForm("https://127.0.0.1:8001/checkHost",
		url.Values{"uuid": {uuid}},
	)
	if err != nil {
		return hosts.Host{}, fmt.Errorf("CheckHost call failed: %s", err)
	}
	if response.StatusCode == 404 {
		return hosts.Host{}, fmt.Errorf("Unknown Host")
	}
	if response.StatusCode != 200 {
		return hosts.Host{}, fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return hosts.Host{}, fmt.Errorf("Could not read response")
	}
	hostResponse := APIHost{}
	err = json.Unmarshal(bodyBytes, &hostResponse)
	if err != nil {
		return hosts.Host{}, err
	}

	return hosts.Host{
		UUID:             hostResponse.UUID,
		ComputerName:     hostResponse.ComputerName,
		Platform:         hostResponse.Platform,
		Version:          hostResponse.Version,
		CurrentDirectory: "/",
	}, nil
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
	if response.StatusCode != 200 {
		return "", fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Could not read response")
	}
	qsResponse := QueryScheduleResponse{}
	err = json.Unmarshal(bodyBytes, &qsResponse)
	if err != nil {
		return "", err
	}
	hosts.AddQueryToCurrentHost(hosts.Query{Name: qsResponse.QueryName, SQL: query})
	return qsResponse.QueryName, nil
}

func ScheduleQueryAndWait(uuid string, query string) ([]map[string]string, error) {
	queryName, err := ScheduleQuery(uuid, query)
	var results = make([]map[string]string, 0)
	if err != nil {
		return results, fmt.Errorf("ScheduleQueryAndWait call failed: %s", err)
	}

	// Wait while the query is pending
	var status string
	for {
		results, status, err = FetchResults(queryName)
		if err != nil || status != "Pending" {
			break
		}
		time.Sleep(time.Second)
		fmt.Printf(".")
		select {
			case <-ctrlcChannel:
				//fmt.Printf("Received Signal: %s", x)
				return results, fmt.Errorf("Waiting Cancelled")
			default:
		}
	}

	fmt.Printf("\n")
	// No need to check error here because return is the same
	return results, err
}

func FetchResults(queryName string) ([]map[string]string, string, error) {
	type ResultsResponse struct {
		Rows   []map[string]string `json:"results"`
		Status string              `json:"status"`
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var client = &http.Client{Transport: tr, Timeout: time.Second * 10}
	resultsResponse := ResultsResponse{}
	response, err := client.PostForm(
		"https://127.0.0.1:8001/fetchResults",
		url.Values{"queryName": {queryName}},
	)

	if err != nil {
		return resultsResponse.Rows, "", fmt.Errorf("FetchResults call failed: %s", err)
	}
	if response.StatusCode == 404 {
		return resultsResponse.Rows, "", fmt.Errorf("Unknown queryName")
	}
	if response.StatusCode != 200 {
		return resultsResponse.Rows, "", fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return resultsResponse.Rows, "", fmt.Errorf("Could not read fetchResults response")
	}

	if err := json.Unmarshal(bodyBytes, &resultsResponse); err != nil {
		return resultsResponse.Rows, "", err
	}

	// Return QueryResultsResponse type (outer caller should check .Status)
	return resultsResponse.Rows, resultsResponse.Status, nil
}
