package osctrl

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/AbGuthrie/goquery/config"

	"github.com/AbGuthrie/goquery/hosts"

	"golang.org/x/crypto/ssh/terminal"
)

// The token type returned by osctrl admin
type TokenResponse struct {
	Token      string `json:"token"`
	Expiration string `json:"expiration"`
}

var token TokenResponse
var cookieJar *cookiejar.Jar
var client *http.Client
var authed bool

var protocol string
var server string
var adminBase string
var apiBase string

func init() {
	authed = false
	protocol = "https"
	adminServer := "osctrl-admin.domain.tld"
	adminBase = fmt.Sprintf("%s://%s", protocol, adminServer)
	apiServer := "osctrl-api.domain.tld"
	apiBase = fmt.Sprintf("%s://%s", protocol, apiServer)

	cookieJar, _ = cookiejar.New(nil)
	debugEnabled := config.GetDebug()
	if debugEnabled {
		fmt.Println("Warning: Debug is enabled, setting InsecureSkipVerify: True for auth request client!")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: debugEnabled,
		},
	}
	client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: cookieJar}
}

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	fmt.Printf("\n")
	return strings.TrimSpace(username), password
}

func Authenticate() error {
	response, err := client.Get(adminBase)

	if err != nil {
		return fmt.Errorf("Couldn't find osctrl-amdin service: %s", err)
	}

	username, _ /*password*/ := credentials()
	// Complete your authentication flow
	fmt.Println("Login Complete")
	fmt.Println("Getting osctrl Token")

	response, err = client.Get(fmt.Sprintf("%s/tokens/%s", adminBase, username))
	if err != nil {
		return fmt.Errorf("Auth call failed: %s", err)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
	}

	contentLength := response.Header["Content-Length"]

	if len(contentLength) > 0 && contentLength[0] == "0" {
		return fmt.Errorf("Server returned no content")
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Could not read token response")
	}

	err = json.Unmarshal(bodyBytes, &token)

	if err != nil {
		return err
	}

	if token.Token == "" {
		return fmt.Errorf("Token returned was empty")
	}

	fmt.Println("Gathered Token Successfully")
	fmt.Printf("%s\n", token.Token)
	authed = true
	return nil
}

func CheckHost(uuid string) (hosts.Host, error) {
	if !authed {
		err := Authenticate()
		if err != nil {
			return hosts.Host{}, err
		}
	}
	type APIHost struct {
		ComputerName   string `json:"Localname"`
		HostIdentifier string `json:"HostIdentifier"`
		Platform       string `json:"Platform"`
		Username       string `json:"Username"`
		UUID           string `json:"UUID"`
		Version        string `json:"OsqueryVersion"`
	}

	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/nodes/%s", apiBase, uuid), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	response, err := client.Do(request)

	if err != nil {
		// Possible Authentication Failure
		return hosts.Host{}, fmt.Errorf("CheckHost call failed: %s", err)
	}

	if response.StatusCode != 200 {
		return hosts.Host{}, fmt.Errorf("Server returned unknown error: %d", response.StatusCode)
	}

	contentLength := response.Header["Content-Length"]

	if len(contentLength) > 0 && contentLength[0] == "0" {
		return hosts.Host{}, fmt.Errorf("Unknown Host")
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
		Username:         hostResponse.Username,
		CurrentDirectory: "/",
	}, nil
}

func ScheduleQuery(uuid string, query string) (string, error) {
	if !authed {
		err := Authenticate()
		if err != nil {
			return "", err
		}
	}
	type QueryScheduleResponse struct {
		QueryName string `json:"query_name"`
	}
	type DistributedQueryRequest struct {
		UUIDs []string `json:"uuid_list"`
		Query string   `json:"query"`
	}

	queryRequest := DistributedQueryRequest{UUIDs: []string{uuid}, Query: query}
	qrJSON, _ := json.Marshal(queryRequest)

	request, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/queries", apiBase), bytes.NewReader(qrJSON))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	response, err := client.Do(request)

	if err != nil {
		Authenticate()
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
		authed = false
		return "", err
	}
	hosts.AddQueryToHost(uuid, hosts.Query{Name: qsResponse.QueryName, SQL: query})
	return qsResponse.QueryName, nil
}

func FetchResults(queryName string) ([]map[string]string, string, error) {
	type ResultsResponse struct {
		Rows   []map[string]string `json:"results"`
		Status string              `json:"status"`
	}
	resultsResponse := ResultsResponse{}

	if !authed {
		err := Authenticate()
		if err != nil {
			return resultsResponse.Rows, "", err
		}
	}

	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/queries/%s", apiBase, queryName), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	response, err := client.Do(request)

	if err != nil {
		Authenticate()
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
		authed = false
		return resultsResponse.Rows, "", err
	}

	// Return QueryResultsResponse type (outer caller should check .Status)
	return resultsResponse.Rows, resultsResponse.Status, nil
}
