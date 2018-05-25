package server

import (
	"net/http"
	"fmt"
)

// Info handles http request on
func Info(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info request"))
		return fmt.Errorf("incorrect method for info request: use GET method")
	}
	var req SearchRequest
	err := req.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect list request"))
		return err
	}
	return nil
}

func List(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for list request"))
		return fmt.Errorf("incorrect method for list request: use GET method")
	}
	var req SearchRequest
	err := req.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect list request"))
		return err
	}
	return nil
}
