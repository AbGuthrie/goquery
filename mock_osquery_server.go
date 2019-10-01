package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Query struct {
	Query  string
	Name   string
	Result string
}

var ENROLL_SECRET string
var enrolledHosts map[string]string
var queryMap map[string]map[string]Query

// API Request Struct
type apiRequest struct {
	NodeKey string `json:"node_key"`
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

// Begin osquery API endpoints
func enroll(w http.ResponseWriter, r *http.Request) {
	type enrollPlatformInfo struct {
		UUID string `json:"uuid"`
	}
	type enrollBody struct {
		EnrollSecret string             `json:"enroll_secret"`
		PlatformInfo enrollPlatformInfo `json:"platform_info"`
	}

	parsedBody := enrollBody{}
	jsonBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(jsonBytes, &parsedBody)
	if err != nil {
		fmt.Printf("Error decoding request JSON: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if parsedBody.EnrollSecret != ENROLL_SECRET {
		fmt.Printf("Host provided incorrrect secret: %s\n", parsedBody.EnrollSecret)
		fmt.Fprintf(w, "{\"node_invalid\" : true}")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeKey := randomString(32)
	fmt.Fprintf(w, "{\"node_key\" : \"%s\"}", nodeKey)
	enrolledHosts[nodeKey] = parsedBody.PlatformInfo.UUID
	queryMap[nodeKey] = make(map[string]Query)
	fmt.Printf("Enrolled a new host with node_key: %s\n", nodeKey)
}

func isNodeKeyEnrolled(ar apiRequest) bool {
	if _, ok := enrolledHosts[ar.NodeKey]; !ok {
		return false
	}
	return true
}

func httpRequestToAPIRequest(r *http.Request) (apiRequest, error) {
	parsedRequest := apiRequest{}
	jsonBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Could not read the body of an API request\n")
		return apiRequest{}, err
	}
	err = json.Unmarshal(jsonBytes, &parsedRequest)
	if err != nil {
		return apiRequest{}, err
	}
	return parsedRequest, nil
}

func config(w http.ResponseWriter, r *http.Request) {
	// This server is designed to test goquery so we don't push a config
	parsedRequest, err := httpRequestToAPIRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !isNodeKeyEnrolled(parsedRequest) {
		fmt.Fprintf(w, "{\"schedule\":{}, \"node_invalid\" : true}")
		return
	}

	fmt.Fprintf(w, "{\"schedule\":{}, \"node_invalid\" : false}")
}

func log(w http.ResponseWriter, r *http.Request) {
	// This server is designed to test goquery so we don't do anything with the logs
}

func distributedRead(w http.ResponseWriter, r *http.Request) {
	parsedRequest, err := httpRequestToAPIRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !isNodeKeyEnrolled(parsedRequest) {
		fmt.Fprintf(w, "{\"node_invalid\" : true}")
		return
	}

	// The check below should never fail. If it does we've really screwed up
	renderedQueries := ""
	if _, ok := queryMap[parsedRequest.NodeKey]; !ok {
		fmt.Fprintf(w, "{\"node_invalid\" : true}")
		fmt.Printf("This should never occur. A host is enrolled but not configured for distributed\n")
		return
	}
	for name, query := range queryMap[parsedRequest.NodeKey] {
		if query.Result != "" {
			continue
		}
		renderedQueries += fmt.Sprintf("\"%s\" : \"%s\",\n", name, query.Query)
	}

	renderedQueries = strings.TrimRight(renderedQueries, ",")
	fmt.Fprintf(w, "{\"queries\" : {%s}}", renderedQueries)
}

func distributedWrite(w http.ResponseWriter, r *http.Request) {
	parsedRequest, err := httpRequestToAPIRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !isNodeKeyEnrolled(parsedRequest) {
		fmt.Fprintf(w, "{\"node_invalid\" : true}")
		return
	}
}

// End osquery API endpoints

func checkHostExists(requestedUUID string) (string, error) {
	for nodeKey, uuid := range enrolledHosts {
		if uuid == requestedUUID {
			return nodeKey, nil
		}
	}
	return "", errors.New("No such host")
}

// Begin goquery APIs
func checkHost(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")
	fmt.Printf("CheckHost call for: %s", r.FormValue("uuid"))
	if _, ok := checkHostExists(uuid); ok != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

func scheduleQuery(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")
	fmt.Printf("ScheduleQuery call for: %s", uuid)
	nodeKey, err := checkHostExists(uuid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	query := Query{}
	query.Name = randomString(64)
	query.Query = r.FormValue("query")

	queryMap[nodeKey][query.Name] = query
	fmt.Fprintf(w, "{\"queryName\" : %s}", query.Name)
}

func fetchResults(w http.ResponseWriter, r *http.Request) {
	queryName := r.FormValue("queryName")
	fmt.Printf("Fetching Results For: %s", queryName)
	// Yes I know this is really slow. For testing it should be fine
	// but I will fix this architecture later if needed
	// The real solution will be to use a better backing store like postgres
	for _, queries := range queryMap {
		if query, ok := queries[queryName]; ok {
			fmt.Fprintf(w, "%s", query.Result)
			return
		}
	}
}

// End goquery APIs

func main() {
	ENROLL_SECRET = "somepresharedsecret"
	enrolledHosts = make(map[string]string)
	queryMap = make(map[string]map[string]Query)
	// TODO enumerate all required endpoints the osquery server must implement

	// osquery Endpoints
	http.HandleFunc("/enroll", enroll)
	http.HandleFunc("/config", config)
	http.HandleFunc("/log", log)
	http.HandleFunc("/distributedRead", distributedRead)
	http.HandleFunc("/distributedWrite", distributedWrite)

	// goquery Endpoints
	http.HandleFunc("/checkHost", checkHost)
	http.HandleFunc("/scheduleQuery", scheduleQuery)
	http.HandleFunc("/fetchResults", fetchResults)

	http.ListenAndServeTLS(":8001", "server.crt", "server.key", nil)
}
