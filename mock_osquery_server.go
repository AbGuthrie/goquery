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
var enrolled_hosts map[string]string

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
	var parsed_request map[string]interface{}
	request_body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	json.Unmarshal([]byte(request_body), &parsed_request)

	if parsed_request["enroll_secret"] == ENROLL_SECRET {
		node_key := random_string(32)
		fmt.Fprintf(w, "{\"node_key\" : \"%s\"}", node_key)
		enrolled_hosts[node_key] = "unknown_uuid"
		fmt.Printf("Enrolled a new host with node_key: %s\n", node_key)
		return
	}
	fmt.Printf("Host provided incorrrect secret: %s\n", parsed_request["enroll_secret"])
	fmt.Fprintf(w, "{\"node_invalid\" : true}")
}

func config(w http.ResponseWriter, r *http.Request)            {}
func log(w http.ResponseWriter, r *http.Request)               {}
func distributed_read(w http.ResponseWriter, r *http.Request)  {}
func distributed_write(w http.ResponseWriter, r *http.Request) {}

// goquery endpoint functions
func check_host(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("CheckHost call for: %s", r.FormValue("uuid"))
}

func main() {
	ENROLL_SECRET = "somepresharedsecret"
	enrolled_hosts = make(map[string]string)
	// TODO enumerate all required endpoints the osquery server must implement

	// GET status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "%UUID% active")
	})

	// osquery Endpoints
	http.HandleFunc("/enroll", enroll)
	http.HandleFunc("/config", config)
	http.HandleFunc("/log", log)
	http.HandleFunc("/distribute_read", distributed_read)
	http.HandleFunc("/distributed_write", distributed_write)

	// goquery Endpoints
	http.HandleFunc("/CheckHost", check_host)

	http.ListenAndServeTLS(":8001", "server.crt", "server.key", nil)
}
