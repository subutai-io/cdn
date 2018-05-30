package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/subutai-io/agent/log"
)

type Hashes struct {
	Md5    string `json:"md5,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

type OldResult struct {
	FileID        string   `json:"id,omitempty"`
	Owner         []string `json:"owner,omitempty"`
	Name          string   `json:"name,omitempty"`
	Filename      string   `json:"filename,omitempty"`
	Version       string   `json:"version,omitempty"`
	Hash          Hashes   `json:"hash,omitempty"`
	Size          int64    `json:"size,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Date          string   `json:"upload-date-formatted,omitempty"`
	Timestamp     string   `json:"upload-date-timestamp,omitempty"`
	Description   string   `json:"description,omitempty"`
	Architecture  string   `json:"architecture,omitempty"`
	Parent        string   `json:"parent,omitempty"`
	ParentVersion string   `json:"parent-version,omitempty"`
	ParentOwner   string   `json:"parent-owner,omitempty"`
	PrefSize      string   `json:"prefsize,omitempty"`
}

// FileSearch handles the info and list HTTP requests
func FileSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info/list request"))
		return
	}
	request := new(SearchRequest)
	request.InitValidators()
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect info/list request: %v", err)))
		return
	}
	files := request.Retrieve()
	oldFiles := make([]*OldResult, 0)
	for _, file := range files {
		oldFiles = append(oldFiles, file.ConvertToOld())
	}
	log.Info("Retrieve: ", files)
	result, _ := json.Marshal(oldFiles)
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
	request.InitUploaders()
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
