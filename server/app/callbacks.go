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
	log.Info("Received FileSearch request")
	log.Info(r)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info/list request"))
		return
	}
	request := new(SearchRequest)
	request.InitValidators()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect info/list request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	files := request.Retrieve()
	log.Info("Successfully retrieved files")
	oldFiles := make([]*OldResult, 0)
	for _, file := range files {
		oldFiles = append(oldFiles, file.ConvertToOld())
	}
	log.Info("Retrieve: ", oldFiles)
	log.Info("Successfully converted files to old format")
	result, _ := json.Marshal(oldFiles)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
	log.Info("Successfully handled FileSearch request")
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	log.Info("Received upload request")
	log.Info(r)
	if r.Method != "POST" {
		log.Warn("Incorrect method for upload request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect method for upload request")))
		return
	}
	request := new(UploadRequest)
	request.InitUploaders()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		log.Warn("Couldn't parse upload request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect upload request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	err = request.Upload()
	if err != nil {
		log.Warn("Couldn't upload file")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while uploading file: %v", err)))
		return
	}
	log.Info("Successfully uploaded a file: ", request.fileID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(request.fileID))
}
