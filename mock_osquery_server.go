package main

import (
	"fmt"
	"strings"
	"time"
	"math/rand"
	"net/http"
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
	if r.FormValue("enroll_secret") == ENROLL_SECRET {
		node_key := random_string(32)
		fmt.Fprintf(w, "{\"node_key\" : \"%s\"}", node_key)
		enrolled_hosts[node_key] = "unknown_uuid"
		return
	}
	fmt.Fprintf(w, "{\"node_invalid\" : True}")
}

func config(w http.ResponseWriter, r *http.Request) {}
func log(w http.ResponseWriter, r *http.Request) {}
func distributed_read(w http.ResponseWriter, r *http.Request) {}
func distributed_write(w http.ResponseWriter, r *http.Request) {}


// goquery endpoint functions
func check_host(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("CheckHost call for: %s", r.FormValue("uuid"));
}

func main(){
	ENROLL_SECRET = "somepresharedkey"
	// TODO enumerate all required endpoints the osquery server must implement

	// GET status
	http.HandleFunc("/status", func (w http.ResponseWriter, r *http.Request) {
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

	http.ListenAndServe(":8001", nil)
}
