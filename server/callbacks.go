package server

import (
	"net/http"
	"fmt"
)

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
		return err
	}
	return nil
}

func List(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info request"))
		return fmt.Errorf("incorrect method for info request: use GET method")
	}
	var req SearchRequest
	err := req.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}
