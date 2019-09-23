package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var ENROLL_SECRET string
var enrolledHosts map[string]string

func random_string(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String() // E.g. "ExcbsVQs"
}

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

	if parsedBody.EnrollSecret == ENROLL_SECRET {
		node_key := random_string(32)
		fmt.Fprintf(w, "{\"node_key\" : \"%s\"}", node_key)
		enrolledHosts[node_key] = parsedBody.PlatformInfo.UUID
		fmt.Printf("Enrolled a new host with node_key: %s\n", node_key)
		return
	}

	fmt.Printf("Host provided incorrrect secret: %s\n", parsedBody.EnrollSecret)
	fmt.Fprintf(w, "{\"node_invalid\" : true}")
	w.WriteHeader(http.StatusBadRequest)
}

func config(w http.ResponseWriter, r *http.Request)           {}
func log(w http.ResponseWriter, r *http.Request)              {}
func distributedRead(w http.ResponseWriter, r *http.Request)  {}
func distributedWrite(w http.ResponseWriter, r *http.Request) {}

// goquery endpoint functions
func checkHost(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("CheckHost call for: %s", r.FormValue("uuid"))
}

func main() {
	ENROLL_SECRET = "somepresharedsecret"
	enrolledHosts = make(map[string]string)
	// TODO enumerate all required endpoints the osquery server must implement

	// GET status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "%UUID% active")
	})

	// osquery Endpoints
	http.HandleFunc("/enroll", enroll)
	http.HandleFunc("/config", config)
	http.HandleFunc("/log", log)
	http.HandleFunc("/distribute_read", distributedRead)
	http.HandleFunc("/distributedWrite", distributedWrite)

	// goquery Endpoints
	http.HandleFunc("/CheckHost", checkHost)

	http.ListenAndServeTLS(":8001", "server.crt", "server.key", nil)
}
