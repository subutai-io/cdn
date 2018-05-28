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
	var request SearchRequest
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect list request"))
		return
	}
	request.operation = "info"
	files := Retrieve(request)
	return json.Marshal(files)
}

// List handles the HTTP request sent on one of the list endpoints
func List(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for list request"))
		return
	}
	var request SearchRequest
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect list request"))
		return
	}
	request.operation = "list"
	files := Retrieve(request)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(files)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("x"))
	}
	w.Write()
}
