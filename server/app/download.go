package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/subutai-io/agent/log"
)

type DownloadRequest struct {
	FileID   string
	Filename string
	Repo     string
	Token    string
	Tags     string
	Version  string

	downloaders map[string]DownloadFunction
}

func (request *DownloadRequest) ParseRequest(r *http.Request) error {
	escapedPath := strings.Split(r.URL.EscapedPath(), "/")
	request.Repo = escapedPath[3]
	if request.Repo == "apt" {
		request.Filename = escapedPath[4]
	}
	request.FileID = r.URL.Query().Get("FileID")
	if request.FileID == "" {
		log.Warn("FileID is empty")
	}
	request.Token = r.Header.Get("token")
	if len(r.MultipartForm.Value["version"]) > 0 {
		request.Version = r.MultipartForm.Value["version"][0]
	}
	if len(r.MultipartForm.Value["tags"]) > 0 {
		request.Tags = r.MultipartForm.Value["tags"][0]
	}
	return nil
}

type DownloadFunction func() (*Result, string, error)

func (request *DownloadRequest) InitDownloaders() {
	request.downloaders = make(map[string]DownloadFunction)
	request.downloaders["apt"] = request.DownloadApt
	request.downloaders["raw"] = request.DownloadFile
	request.downloaders["template"] = request.DownloadFile
}

func (request *DownloadRequest) ExecRequest() (*Result, string, error) {
	downloader := request.downloaders[request.Repo]
	return downloader()
}

func (request *DownloadRequest) DownloadFile() (*Result, string, error) {
	searchRequest := &SearchRequest{
		FileID:  request.FileID,
		Name:    request.Filename,
		Repo:    request.Repo,
		Token:   request.Token,
		Tags:    request.Tags,
		Version: request.Version,
	}
	list := searchRequest.Retrieve()
	if len(list) == 0 {
		return nil, "", fmt.Errorf("Files not found")
	}
	return list[0], request.Token, nil
}

func (request *DownloadRequest) DownloadApt() (*Result, string, error) {
	searchRequest := &SearchRequest{
		FileID:  request.FileID,
		Name:    request.Filename,
		Repo:    request.Repo,
		Token:   request.Token,
		Tags:    request.Tags,
		Version: request.Version,
	}
	list := searchRequest.Retrieve()
	if len(list) == 0 {
		return nil, "", fmt.Errorf("Files not found")
	}
	return list[0], "", nil
}
