package app

import (
	"net/http"
	"strings"
	"github.com/subutai-io/cdn/db"
	"github.com/boltdb/bolt"
	"fmt"
)

type SearchRequest struct {
	fileID  string
	name    string
	owner   string
	token   string
	tags    string
	version string
	repo    string
}

// ParseRequest takes HTTP request and converts it into Request struct
func (r *SearchRequest) ParseRequest(req *http.Request) (err error) {
	r.fileID = req.URL.Query().Get("id")
	r.name = req.URL.Query().Get("name")
	r.owner = req.URL.Query().Get("owner")
	r.token = req.URL.Query().Get("token")
	r.tags = req.URL.Query().Get("tags")
	r.version = req.URL.Query().Get("version")
	r.repo = strings.Split(req.RequestURI, "/")[3] // Splitting /kurjun/rest/repo/func into ["", "kurjun", "rest", "repo" (index: 3), "func"]
	return
}

// BuildQuery constructs the query out of the existing parameters in SearchRequest
func (r *SearchRequest) BuildQuery() (query map[string]string) {
	if r.fileID != "" {
		query["fileID"] = r.fileID
	}
	if r.name != "" {
		query["name"] = r.name
	}
	if r.owner != "" {
		query["owner"] = r.owner
	}
	if r.token != "" {
		query["token"] = r.token
	}
	if r.tags != "" {
		query["tags"] = r.tags
	}
	if r.version != "" {
		query["version"] = r.version
	}
	if r.repo != "" {
		query["repo"] = r.repo
	}
	return
}

type SearchResult struct {

}

func Retrieve(request SearchRequest) []SearchResult {
	query := request.BuildQuery()
	results, err := Search(query)
	if err != nil {

	}
	return results
}

func GetFileInfo(id string) (info map[string]string, err error) {
	err = db.DB.View(func(tx *bolt.Tx) error {
		file := tx.Bucket(db.MyBucket).Bucket([]byte(id))
		if file == nil {
			return fmt.Errorf("file %s not found", id)
		}

	})
	if err != nil {
		return nil, err
	}

}

func Search(query map[string]string) ([]SearchResult, error) {
	db.DB.View(func(tx *bolt.Tx) error {
		files := tx.Bucket(db.MyBucket)
		files.ForEach(func(k, v []byte) error {

			return nil
		})
		return nil
	})
}