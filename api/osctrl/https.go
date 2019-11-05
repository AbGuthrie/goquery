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

	"github.com/AbGuthrie/goquery/api/models"
	"github.com/AbGuthrie/goquery/config"

	"github.com/AbGuthrie/goquery/hosts"

	"golang.org/x/crypto/ssh/terminal"
)

// The token type returned by osctrl admin
type tokenResponse struct {
	Token      string `json:"token"`
	Expiration string `json:"expiration"`
}

type osctrlAPI struct {
	Token     tokenResponse
	CookieJar *cookiejar.Jar
	Client    *http.Client
	Authed    bool

	Protocol  string
	Server    string
	AdminBase string
	APIBase   string
}

var instance osctrlAPI

// Initialize creates and returns an api implementation that implements the models.GoQueryAPI interface
// can easily be parameterized with flags passed from main via the config.json
func Initialize() (models.GoQueryAPI, error) {
	adminServer := "osctrl-admin.domain.tld"
	protocol := "https"
	apiServer := "osctrl-api.domain.tld"

	instance = osctrlAPI{
		Authed:    false,
		Protocol:  protocol,
		AdminBase: fmt.Sprintf("%s://%s", protocol, adminServer),
		APIBase:   fmt.Sprintf("%s://%s", protocol, apiServer),
	}

	instance.CookieJar, _ = cookiejar.New(nil)
	debugEnabled := config.GetDebug()
	if debugEnabled {
		fmt.Println("Warning: Debug is enabled, setting InsecureSkipVerify: True for auth request client!")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: debugEnabled,
		},
	}
	instance.Client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: instance.CookieJar}

	return &instance, nil
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

func (instance *osctrlAPI) authenticate() error {
	response, err := instance.Client.Get(instance.AdminBase)

	if err != nil {
		return fmt.Errorf("Couldn't find osctrl-amdin service: %s", err)
	}

	username, _ /*password*/ := credentials()
	// Complete your authentication flow
	fmt.Println("Login Complete")
	fmt.Println("Getting osctrl Token")

	response, err = instance.Client.Get(fmt.Sprintf("%s/tokens/%s", instance.AdminBase, username))
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

	err = json.Unmarshal(bodyBytes, &instance.Token)

	if err != nil {
		return err
	}

	if instance.Token.Token == "" {
		return fmt.Errorf("Token returned was empty")
	}

	fmt.Println("Gathered Token Successfully")
	fmt.Printf("%s\n", instance.Token.Token)
	instance.Authed = true
	return nil
}

func (instance *osctrlAPI) CheckHost(uuid string) (hosts.Host, error) {
	if !instance.Authed {
		err := instance.authenticate()
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

	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/nodes/%s", instance.APIBase, uuid), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", instance.Token.Token))
	response, err := instance.Client.Do(request)

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

func (instance *osctrlAPI) ScheduleQuery(uuid string, query string) (string, error) {
	if !instance.Authed {
		err := instance.authenticate()
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

	request, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/queries", instance.APIBase), bytes.NewReader(qrJSON))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", instance.Token.Token))
	response, err := instance.Client.Do(request)

	if err != nil {
		instance.authenticate()
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
		instance.Authed = false
		return "", err
	}
	hosts.AddQueryToHost(uuid, hosts.Query{Name: qsResponse.QueryName, SQL: query})
	return qsResponse.QueryName, nil
}

func (instance *osctrlAPI) FetchResults(queryName string) ([]map[string]string, string, error) {
	type ResultsResponse struct {
		Rows   []map[string]string `json:"results"`
		Status string              `json:"status"`
	}
	resultsResponse := ResultsResponse{}

	if !instance.Authed {
		err := instance.authenticate()
		if err != nil {
			return resultsResponse.Rows, "", err
		}
	}

	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/queries/%s", instance.APIBase, queryName), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", instance.Token.Token))
	response, err := instance.Client.Do(request)

	if err != nil {
		instance.authenticate()
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
		instance.Authed = false
		return resultsResponse.Rows, "", err
	}

	// Return QueryResultsResponse type (outer caller should check .Status)
	return resultsResponse.Rows, resultsResponse.Status, nil
}
