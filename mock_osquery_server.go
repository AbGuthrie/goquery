package main

import (
	"fmt"
	"net/http"
)

func main(){
	// TODO enumerate all required endpoints the osquery server must implement

	// GET status
	http.HandleFunc("/status", func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "%UUID% active")
	})

	http.ListenAndServe(":8001", nil)
}