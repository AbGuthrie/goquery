package api

import (
	"bufio"
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
	"syscall"
	"time"

	"github.com/AbGuthrie/goquery/config"

	"github.com/AbGuthrie/goquery/hosts"

	"golang.org/x/crypto/ssh/terminal"
)

var ctrlcChannel (chan os.Signal)
var token string
var cookieJar *cookiejar.Jar
var client *http.Client
var authed bool

func init() {
	authed = false
	ctrlcChannel = make(chan os.Signal, 1)
	signal.Notify(ctrlcChannel, os.Interrupt)
	cookieJar, _ = cookiejar.New(nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.GetConfig().DebugEnabled,
		},
	}
	client = &http.Client{Transport: tr, Timeout: time.Second * 10, Jar: cookieJar}
	err := Authenticate()
	if err != nil {
		fmt.Printf("Could not authenticate with the backend\n")
	}
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

func extractSSORequest(response *http.Response) (string, string) {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", ""
	}
	bodyStr := string(bodyBytes)
	// Hacky Extracts
	loc := strings.Index(bodyStr, "name=\"SAMLRequest\"")
	endLoc := strings.Index(bodyStr[loc+26:], "\" ")
	samlRequest := bodyStr[loc+26 : loc+26+endLoc]

	loc = strings.Index(bodyStr, "name=\"RelayState\"")
	endLoc = strings.Index(bodyStr[loc+17:], "\" ")
	relayState := bodyStr[loc+25 : loc+17+endLoc]
	return samlRequest, relayState
}

func extractSSOResponse(response *http.Response) (string, string, error) {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}
	bodyStr := string(bodyBytes)
	if strings.Index(bodyStr, "Invalid username or password") != -1 {
		return "", "", fmt.Errorf("Credential Failure")
	}
	// Hacky Extracts
	loc := strings.Index(bodyStr, "name=\"SAMLResponse\"")
	endLoc := strings.Index(bodyStr[loc+27:], "\" ")
	ssoResponse := bodyStr[loc+27 : loc+27+endLoc]

	loc = strings.Index(bodyStr, "name=\"RelayState\"")
	endLoc = strings.Index(bodyStr[loc+17:], "\" ")
	relayState := bodyStr[loc+25 : loc+17+endLoc]
	return ssoResponse, relayState, nil
}

func Authenticate() error {
	response, err := client.Get("https://localhost:8001/checkHost")

	if err != nil {
		return fmt.Errorf("Authentication failed: %s", err)
	}

	fmt.Printf("Authenticating with backend...\n")
	ssoRequest, relayState := extractSSORequest(response)
	username, password := credentials()

	// TODO This should be an HTTPS endpoint and should use the global client
	var httpClient = &http.Client{Timeout: time.Second * 10, Jar: cookieJar}
	response, err = httpClient.PostForm("http://localhost:8002/sso",
		url.Values{"SAMLRequest": {ssoRequest}, "RelayState": {relayState}, "user": {username}, "password": {password}},
	)
	if err != nil {
		return err
	}

	samlResponse, relayState, err := extractSSOResponse(response)
	if err != nil {
		return err
	}

	response, err = client.PostForm("https://localhost:8001/saml/acs",
		url.Values{"SAMLResponse": {samlResponse}, "RelayState": {relayState}},
	)

	if err != nil {
		return fmt.Errorf("Authentication failed: %s", err)
	}
	fmt.Printf("Authentication Complete\n")
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
		UUID           string `json:"UUID"`
		ComputerName   string `json:"ComputerName"`
		HostIdentifier string `json:"HostIdentifier"`
		Platform       string `json:"Platform"`
		Version        string `json:"Version"`
	}

	response, err := client.PostForm("https://localhost:8001/checkHost",
		url.Values{"uuid": {uuid}},
	)
	if err != nil {
		// Possible Authentication Failure
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
		// Probable authentication failure
		authed = false
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
	if !authed {
		err := Authenticate()
		if err != nil {
			return "", err
		}
	}
	type QueryScheduleResponse struct {
		QueryName string `json:"queryName"`
	}

	response, err := client.PostForm("https://localhost:8001/scheduleQuery",
		url.Values{
			"uuid":  {uuid},
			"query": {query}},
	)
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
	resultsResponse := ResultsResponse{}

	if !authed {
		err := Authenticate()
		if err != nil {
			return resultsResponse.Rows, "", err
		}
	}

	response, err := client.PostForm(
		"https://localhost:8001/fetchResults",
		url.Values{"queryName": {queryName}},
	)

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
