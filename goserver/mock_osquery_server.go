package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Query struct {
	Query    string
	Name     string
	Complete bool
	Result   json.RawMessage `json:"results"`
	Status   string          `json:"status"`
}

type Host struct  {
	UUID string
	ComputerName string
	HostIdentifier string
	Platform string
	Version string
}

var ENROLL_SECRET string

// Maps Node Key -> UUID
var enrolledHosts map[string]Host

// Maps Node Key -> Map of Query Name -> Query struct
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
	type enrollSystemInfo struct {
		UUID string `json:"uuid"`
		ComputerName string `json:"computer_name"`
	}
	type hostDetailsBody struct {
		SystemInfo enrollSystemInfo `json:"system_info"`
	}
	type enrollBody struct {
		EnrollSecret string          `json:"enroll_secret"`
		HostIdentifier string        `json:"host_identifier"`
		HostDetails  hostDetailsBody `json:"host_details"`
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
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"node_invalid\" : true}")
		return
	}
	nodeKey := randomString(32)
	fmt.Fprintf(w, "{\"node_key\" : \"%s\"}", nodeKey)
	// The configuration is overriding the host_identifier with something else so we
	// should definitely use that for indexing
	if parsedBody.HostIdentifier != "" {
		enrolledHosts[nodeKey] = Host{UUID: parsedBody.HostIdentifier, ComputerName: parsedBody.HostDetails.SystemInfo.ComputerName}
	} else {
		enrolledHosts[nodeKey] = Host{UUID: parsedBody.HostDetails.SystemInfo.UUID, ComputerName: parsedBody.HostDetails.SystemInfo.ComputerName}
	}
	queryMap[nodeKey] = make(map[string]Query)
	fmt.Printf("Enrolled a host (%s) with node_key: %s\n", enrolledHosts[nodeKey].UUID, nodeKey)
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
		if query.Complete {
			continue
		}
		renderedQueries += fmt.Sprintf("\"%s\" : \"%s\",", name, query.Query)
	}

	renderedQueries = strings.TrimRight(renderedQueries, ",")
	fmt.Fprintf(w, "{\"queries\" : {%s}}", renderedQueries)
}

func distributedWrite(w http.ResponseWriter, r *http.Request) {
	jsonBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Could not read body: %s\n", err)
		return
	}

	type distributedResponse struct {
		Queries  map[string]json.RawMessage `json:"queries"`
		Statuses map[string]int             `json:"statuses"`
		NodeKey  string                     `json:"node_key"`
	}

	// Decode request body, but don't bother decoding the query results
	// These should be opaquely passed along when asked for
	responseParsed := distributedResponse{}
	if err := json.Unmarshal(jsonBytes, &responseParsed); err != nil {
		fmt.Printf("Could not parse body: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !isNodeKeyEnrolled(apiRequest{NodeKey: responseParsed.NodeKey}) {
		fmt.Fprintf(w, "{\"node_invalid\" : true}")
		fmt.Printf("The host sending results is not enrolled\n")
		return
	}

	type responseQuery struct {
		Rows     json.RawMessage
		Status   string
		SQLQuery string
	}
	responses := make(map[string]*responseQuery)
	for queryName, resultsRaw := range responseParsed.Queries {
		sqlQuery := queryMap[responseParsed.NodeKey][queryName].Query
		responses[queryName] = &responseQuery{
			SQLQuery: sqlQuery,
			Rows:     resultsRaw,
		}
	}
	for queryName, statusCode := range responseParsed.Statuses {
		if statusCode == 0 {
			responses[queryName].Status = "Complete"
		} else {
			responses[queryName].Status = fmt.Sprintf("Status Code %d", statusCode)
		}
	}

	for queryName, response := range responses {
		queryMap[responseParsed.NodeKey][queryName] = Query{
			Query:    response.SQLQuery,
			Name:     queryName,
			Complete: true,
			Result:   response.Rows,
			Status:   response.Status,
		}
		fmt.Printf("Received and set query results for %s\n", queryName)
	}
}

// End osquery API endpoints

func checkHostExists(requestedUUID string) (string, error) {
	for nodeKey, host := range enrolledHosts {
		if host.UUID == requestedUUID {
			return nodeKey, nil
		}
	}
	return "", errors.New("No such host")
}

// Begin goquery APIs
func checkHost(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")
	fmt.Printf("CheckHost call for: %s\n", r.FormValue("uuid"))
	var nodeKey string
	var ok error
	if nodeKey, ok = checkHostExists(uuid); ok != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	renderedHost, err := json.Marshal(enrolledHosts[nodeKey])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", renderedHost)
}

func scheduleQuery(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("uuid")
	sentQuery := r.FormValue("query")

	fmt.Printf("ScheduleQuery call for: %s with query: %s\n", uuid, sentQuery)
	nodeKey, err := checkHostExists(uuid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	query := Query{
		Name:   randomString(64),
		Query:  sentQuery,
		Status: "Pending",
	}

	queryMap[nodeKey][query.Name] = query
	fmt.Fprintf(w, "{\"queryName\" : \"%s\"}", query.Name)
}

func fetchResults(w http.ResponseWriter, r *http.Request) {
	queryName := r.FormValue("queryName")
	fmt.Printf("Fetching Results For: %s\n", queryName)
	// Yes I know this is really slow. For testing it should be fine
	// but I will fix this architecture later if needed
	// The real solution will be to use a better backing store like postgres
	for _, queries := range queryMap {
		if query, ok := queries[queryName]; ok {
			bytes, err := json.MarshalIndent(&query, "", "\t")
			if err != nil {
				fmt.Printf("Could not encode query result: %s\n", err)
				fmt.Fprintf(w, "Could not encode query result: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(bytes)
			return
		}
	}
}

// End goquery APIs

func main() {
	ENROLL_SECRET = "somepresharedsecret"
	enrolledHosts = make(map[string]Host)
	queryMap = make(map[string]map[string]Query)

	// Set up flags for certs
	serverCrt := flag.String("server_cert", "certs/example_server.crt", "Location of a certificate to use")
	serverKey := flag.String("server_key", "certs/example_server.key", "Location of key for certificate")

	flag.Parse()

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

	fmt.Printf("Starting test goquery/osquery backend...\n")
	fmt.Printf("Server Cert Path: %s\n", *serverCrt)
	fmt.Printf("Server Key Path:  %s\n", *serverKey)

	err := http.ListenAndServeTLS(":8001", *serverCrt, *serverKey, nil)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
