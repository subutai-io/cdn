package app

import (
	"github.com/subutai-io/cdn/db"
	"fmt"
	"net/http"
	"strings"
	"os"
	"github.com/subutai-io/cdn/config"
)

type DeleteRequest struct {
	FileID string `json:"FileID,omitempty"` // files' UUID (or MD5, or filename)
	Owner  string `json:"Owner,omitempty"`  // owner of file
	Token  string `json:"Token,omitempty"`  // user's token
	Repo   string `json:"Repo,omitempty"`   // files' repository - either "apt", "raw", or "template"
}

func (request *DeleteRequest) ValidateRequest() error {
	if request.FileID == "" {
		return fmt.Errorf("file ID wasn't provided")
	}
	if request.Token == "" || db.TokenOwner(request.Token) == "" {
		return fmt.Errorf("provided invalid token")
	}
	return nil
}

func (request *DeleteRequest) ParseRequest(r * http.Request) error {
	request.FileID = r.URL.Query().Get("id")
	request.Token = r.URL.Query().Get("token")
	request.Owner = db.TokenOwner(request.Token)
	request.Repo = strings.Split(r.URL.EscapedPath(), "/")[3]
	return request.ValidateRequest()
}

func (request *DeleteRequest) Delete() error {
	searchRequest := &SearchRequest{
		FileID:    request.FileID,
		Owner:     request.Owner,
		Token:     request.Token,
		Repo:      request.Repo,
		Operation: "list",
	}
	list := searchRequest.Retrieve()
	if len(list) == 0 {
		return fmt.Errorf("no files found")
	}
	DeleteDB(list[0])
	DeleteFS(list[0])
	return nil
}

func DeleteFS(result *Result) {
	if CountDB(result) == 0 {
		if result.Repo != "apt" {
			os.Remove(config.ConfigurationStorage.Path + result.Md5)
		} else {
			os.Remove(config.ConfigurationStorage.Path + result.Filename)
		}
	}
}
