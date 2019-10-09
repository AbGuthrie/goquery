package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/AbGuthrie/goquery/hosts"
)

var ctrlcChannel (chan os.Signal)
var token string
var cookieJar *cookiejar.Jar

func init() {
	ctrlcChannel = make(chan os.Signal, 1)
	signal.Notify(ctrlcChannel, os.Interrupt)
	cookieJar, _ = cookiejar.New(nil)
	Authenticate()
}

func extractSSORequest(response *http.Response) (string, string) {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", ""
	}
	bodyStr := string(bodyBytes)
	// Hacky Extracts
	loc := strings.Index(bodyStr, "name=\"SAMLRequest\"")
	endLoc := strings.Index(bodyStr[loc+26:], "\" ")
	samlRequest := bodyStr[loc+26:loc+26+endLoc]

	loc = strings.Index(bodyStr, "name=\"RelayState\"")
	endLoc = strings.Index(bodyStr[loc+17:], "\" ")
	relayState := bodyStr[loc+25:loc+17+endLoc]
	return samlRequest, relayState
}

func extractSSOResponse(response *http.Response) (string, string) {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", ""
	}
	bodyStr := string(bodyBytes)
	// Hacky Extracts
	loc := strings.Index(bodyStr, "name=\"SAMLResponse\"")
	endLoc := strings.Index(bodyStr[loc+27:], "\" ")
	ssoResponse := bodyStr[loc+27:loc+27+endLoc]

	loc = strings.Index(bodyStr, "name=\"RelayState\"")
	endLoc = strings.Index(bodyStr[loc+17:], "\" ")
	relayState := bodyStr[loc+25:loc+17+endLoc]
	return ssoResponse, relayState
}

func Authenticate() error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: cookieJar}
	response, err := client.PostForm("https://localhost:8001/checkHost",
		url.Values{"uuid": {"00000000-0000-0000-0000-000000000000"}},
	)

	if err != nil {
		return fmt.Errorf("Authentication failed: %s", err)
	}

	fmt.Printf("Authenticating with backend...\n")

	ssoRequest, relayState := extractSSORequest(response)
	//fmt.Printf("SSORequest: %s\n", ssoRequest)
	//fmt.Printf("RelayState: %s\n", relayState)


	var httpClient = &http.Client{Timeout: time.Second * 10, Jar: cookieJar}
	response, err = httpClient.PostForm("http://localhost:8002/sso",
		url.Values{"SAMLRequest": {ssoRequest}, "RelayState": {relayState}, "user" : {"alice"}, "password" : {"hunter2"}},
	)
	if err != nil {
		return err
	}
	samlResponse, relayState := extractSSOResponse(response)
	//fmt.Printf("SAMLResponse: %s\n", samlResponse)
	//fmt.Printf("RelayState: %s\n", relayState)

	response, err = client.PostForm("https://localhost:8001/saml/acs",
		url.Values{"SAMLResponse": {samlResponse}, "RelayState" : {relayState}},
	)

	//fmt.Printf("%s\n", response.Header)

	if err != nil {
		return fmt.Errorf("Authentication failed: %s", err)
	}
	fmt.Printf("Authentication Complete\n")
	return nil
}

func CheckHost(uuid string) (hosts.Host, error) {
	type APIHost struct {
		UUID           string `json:"UUID"`
		ComputerName   string `json:"ComputerName"`
		HostIdentifier string `json:"HostIdentifier"`
		Platform       string `json:"Platform"`
		Version        string `json:"Version"`
	}

	//TODO Remove later
	//err := Authenticate()
	//return hosts.Host{}, err

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: cookieJar}
	response, err := client.PostForm("https://localhost:8001/checkHost",
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
	var client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: cookieJar}
	response, err := client.PostForm("https://localhost:8001/scheduleQuery",
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

	var client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: cookieJar}
	resultsResponse := ResultsResponse{}
	response, err := client.PostForm(
		"https://localhost:8001/fetchResults",
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
