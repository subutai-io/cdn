package app

import (
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
	request.FileID = r.URL.Query().Get("FileID")
	if request.FileID == "" {
		log.Warn("FileID is empty")
	}
	escapedPath := strings.Split(r.URL.EscapedPath(), "/")
	request.Repo = escapedPath[3]
	request.Token = r.Header.Get("token")
	if len(r.MultipartForm.Value["version"]) > 0 {
		request.Version = r.MultipartForm.Value["version"][0]
	}
	if len(r.MultipartForm.Value["tags"]) > 0 {
		request.Tags = r.MultipartForm.Value["tags"][0]
	}
	return nil
}

type DownloadFunction func() error

func (request *DownloadRequest) InitDownloaders() {
	request.downloaders = make(map[string]DownloadFunction)
	request.downloaders["apt"] = request.DownloadApt
	request.downloaders["raw"] = request.DownloadFile
	request.downloaders["template"] = request.DownloadFile
}

func (request *DownloadRequest) ExecRequest() error {
	downloader := request.downloaders[request.Repo]
	return downloader()
}

func (request *DownloadRequest) DownloadFile() error {
	return nil
}

func (request *DownloadRequest) DownloadApt() error {
	return nil
}
