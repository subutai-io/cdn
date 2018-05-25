package server

import (
	"net/http"
	"strings"
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

func Retrieve() {

}