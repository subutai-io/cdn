package app

import (
	"github.com/subutai-io/cdn/db"
	"fmt"
	"net/http"
	"strings"
	"github.com/boltdb/bolt"
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
		FileID: request.FileID,
		Owner:  request.Owner,
		Token:  request.Token,
		Repo:   request.Repo,
	}
	searchRequest.InitValidators()
	list := searchRequest.Retrieve()
	if len(list) == 0 {
		return fmt.Errorf("no files found")
	}
	if len(list) > 1 {
		return fmt.Errorf("more than one file exist")
	}
	DeleteDB(list[0])
	DeleteFS(list[0])
	return nil
}

func CountFile(md5 string) int {
	answer := 0
	db.DB.View(func(tx *bolt.Tx) error {
		myBucket := tx.Bucket(db.MyBucket)
		myBucket.ForEach(func(k, v []byte) error {
			file := myBucket.Bucket(k)
			if hash := file.Bucket([]byte("hash")); hash != nil {
				md5Hash := string(hash.Get([]byte("md5")))
				if md5Hash == md5 {
					answer++
				}
			}
			return nil
		})
		return nil
	})
	return answer
}

func DeleteFS(result *Result) {
	if CountFile(result.Md5) == 1 {
		os.Remove(config.ConfigurationStorage.Path + result.Filename)
	}
}
