package mock

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
	"strings"
	"syscall"
	"time"

	"github.com/AbGuthrie/goquery/api/models"
	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"

	"golang.org/x/crypto/ssh/terminal"
)

type mockAPI struct {
	Token     string
	CookieJar *cookiejar.Jar
	Client    *http.Client
	Authed    bool
}

var instance mockAPI

// Intialize creates and returns an api implementation that implements the models.GoQueryAPI interface
// can easily be parameterized with flags passed from main via the config.json
func Intialize() (models.GoQueryAPI, error) {
	instance = mockAPI{
		Authed: false,
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

func extractSSORequest(response *http.Response) (string, string) {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", ""
	}
	bodyStr := string(bodyBytes)
	// Hacky Extracts
	loc := strings.Index(bodyStr, "name=\"SAMLRequest\"")
	if loc == -1 {
		return "", ""
	}
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
	if config.GetDebug() {
		fmt.Printf("ssoResponse: %s\n", bodyStr)
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

func (instance *mockAPI) authenticate() error {
	response, err := instance.Client.Get("https://localhost:8001/checkHost")
	if err != nil {
		return fmt.Errorf("Authentication failed: %s", err)
	}

	fmt.Printf("Authenticating with backend...\n")
	ssoRequest, relayState := extractSSORequest(response)

	if ssoRequest == "" && relayState == "" {
		// Looks like the user is already authed, there was no SAML data
		instance.Authed = true
		return nil
	}

	if config.GetDebug() {
		fmt.Printf("ssoRequest: %s\nrelayState: %s\n", ssoRequest, relayState)
	}

	username, password := credentials()

	response, err = instance.Client.PostForm("http://localhost:8002/sso",
		url.Values{"SAMLRequest": {ssoRequest}, "RelayState": {relayState}, "user": {username}, "password": {password}},
	)
	if err != nil {
		return err
	}

	samlResponse, relayState, err := extractSSOResponse(response)
	if config.GetDebug() {
		fmt.Printf("ssoResponse: %s\nrelayState: %s\n", samlResponse, relayState)
	}

	if err != nil {
		return err
	}

	response, err = instance.Client.PostForm("https://localhost:8001/saml/acs",
		url.Values{"SAMLResponse": {samlResponse}, "RelayState": {relayState}},
	)

	if config.GetDebug() {
		fmt.Printf("samlResponse: %s\nrelayState: %s\n", response, relayState)
	}

	if err != nil {
		return fmt.Errorf("Authentication failed: %s", err)
	}

	fmt.Printf("Authentication Complete\n")
	instance.Authed = true
	return nil
}

func (instance *mockAPI) CheckHost(uuid string) (hosts.Host, error) {
	if !instance.Authed {
		err := instance.authenticate()
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

	response, err := instance.Client.PostForm("https://localhost:8001/checkHost",
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
		if config.GetDebug() {
			fmt.Printf("Returned Body: %s\n", string(bodyBytes))
		}
		// Probable authentication failure
		instance.Authed = false
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

func (instance *mockAPI) ScheduleQuery(uuid string, query string) (string, error) {
	if !instance.Authed {
		err := instance.authenticate()
		if err != nil {
			return "", err
		}
	}
	type QueryScheduleResponse struct {
		QueryName string `json:"queryName"`
	}

	response, err := instance.Client.PostForm("https://localhost:8001/scheduleQuery",
		url.Values{
			"uuid":  {uuid},
			"query": {query}},
	)
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

func (instance *mockAPI) FetchResults(queryName string) ([]map[string]string, string, error) {
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

	response, err := instance.Client.PostForm(
		"https://localhost:8001/fetchResults",
		url.Values{"queryName": {queryName}},
	)

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
