package app

import (
	"encoding/json"
	"net/http"
	"fmt"
)

// FileSearch handles the info and list HTTP requests
func FileSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info/list request"))
		return
	}
	request := new(SearchRequest)
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect info/list request: %v", err)))
		return
	}
	files := request.Retrieve()
	result, _ := json.Marshal(files)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect method for upload request")))
		return
	}
	request := new(UploadRequest)
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect upload request: %v", err)))
		return
	}
	err = request.Upload()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while uploading file: %v", err)))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}
