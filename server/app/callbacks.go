package app

import (
	"net/http"
)

// Info handles the HTTP request sent on one of the info endpoints
func Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info request"))
		return
	}
	var req SearchRequest
	err := req.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect list request"))
		return
	}
}

// List handles the HTTP request sent on one of the list endpoints
func List(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for list request"))
		return
	}
	var req SearchRequest
	err := req.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect list request"))
		return
	}
}
