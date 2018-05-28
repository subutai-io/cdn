package app

import (
	"fmt"
	"time"
	"strings"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/blang/semver"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
)

type SearchRequest struct {
	fileID    string // files' UUID (or MD5)
	owner     string // files' owner username
	name      string // files' name within CDN
	repo      string // files' repository - either "apt", "raw", or "template"
	version   string // files' version
	tags      string // files' tags in format: "tag1,tag2,tag3"
	token     string // user's token
	operation string // operation type requested
}

// ParseRequest takes HTTP request and converts it into Request struct
func (request *SearchRequest) ParseRequest(httpRequest *http.Request) (err error) {
	request.fileID = httpRequest.URL.Query().Get("id")
	request.name = httpRequest.URL.Query().Get("name")
	request.owner = httpRequest.URL.Query().Get("owner")
	request.repo = strings.Split(httpRequest.RequestURI, "/")[3] // Splitting /kurjun/rest/repo/func into ["", "kurjun", "rest", "repo" (index: 3), "func"]
	request.version = httpRequest.URL.Query().Get("version")
	request.tags = httpRequest.URL.Query().Get("tags")
	request.token = httpRequest.URL.Query().Get("token")
	return
}

type validateFunc func(SearchRequest) error

var (
	validators map[string]validateFunc
)

func InitValidators() {
	validators["info"] = ValidateInfo
	validators["list"]	= ValidateList
}

func (request SearchRequest) ValidateRequest() error {
	validator := validators[request.operation]
	return validator(request)
}

func ValidateInfo(request SearchRequest) error {
	return nil
}

func ValidateList(request SearchRequest) error {
	return nil
}

// BuildQuery constructs the query out of the existing parameters in SearchRequest
func (r *SearchRequest) BuildQuery() (query map[string]string) {
	if r.fileID != "" {
		query["fileID"] = r.fileID
	}
	if r.owner != "" {
		query["owner"] = r.owner
	}
	if r.name != "" {
		query["name"] = r.name
	}
	if r.repo != "" {
		query["repo"] = r.repo
	}
	if r.version != "" {
		if r.version == "latest" {
			r.version = ""
		}
		query["version"] = r.version
	}
	if r.tags != "" {
		query["tags"] = r.tags
	}
	return
}

type SearchResult struct {
	fileID        string `json:"id,omitempty"`            // file's UUID (or MD5)
	owner         string `json:"owner,omitempty"`         // file owner's username
	name          string `json:"name,omitempty"`          // file's name within CDN
	repo          string `json:"repo,omitempty"`          // file's repository - either "apt", "raw", or "template"
	version       string `json:"version,omitempty"`       // file's version
	tags          string `json:"tags,omitempty"`          // file's tags in format: "tag1,tag2,tag3"
	scope         string `json:"scope,omitempty"`         // file's availibility scope - public/private and users with whom it was shared
	md5           string `json:"md5,omitempty"`           // file's MD5
	sha256        string `json:"sha256,omitempty"`        // file's SHA256
	date          string `json:"date,omitempty"`          // file's upload date
	size          string `json:"size,omitempty"`          // file's size in bytes
	parent        string `json:"parent,omitempty"`        // template's parent template
	parentOwner   string `json:"parentOwner,omitempty"`   // parent template owner's username
	parentVersion string `json:"parentVersion,omitempty"` // parent template's version
	prefSize      string `json:"prefSize,omitempty"`      // template's preffered size
	architecture  string `json:"architecture,omitempty"`  // template's architecture
}

type filterFunc func(map[string]string, []SearchResult) []SearchResult

var (
	filters map[string]filterFunc
)

func InitFilters() {
	filters["info"] = FilterInfo
	filters["list"]	= FilterList
}

func Retrieve(request SearchRequest) []SearchResult {
	query := request.BuildQuery()
	files, err := Search(query)
	if err != nil {
		log.Error("retrieve couldn't search the query %+v", query)
		return nil
	}
	filter := filters[request.operation]
	results := filter(query, files)
	return results
}

func FilterInfo(query map[string]string, files []SearchResult) (result []SearchResult) {
	queryVersion, _ := semver.Make(query["version"])
	for _, file := range files {
		fileVersion, _ := semver.Make(file.version)
		if query["version"] == "" {
			if fileVersion.GTE(queryVersion) {
				queryVersion = fileVersion
				result = []SearchResult{file}
			} else if fileVersion.EQ(queryVersion) && len(result) > 0 {
				resultDate, _ := time.Parse(time.RFC3339, result[0].date)
				fileDate, _ := time.Parse(time.RFC3339, file.date)
				if resultDate.After(fileDate) {
					result = []SearchResult{file}
				}
			}
		} else {
			if queryVersion.EQ(fileVersion) {
				result = []SearchResult{file}
			}
		}
	}
	return
}

func FilterList(query map[string]string, files []SearchResult) (results []SearchResult) {
	owner := query["owner"]
	token := query["token"]
	if owner == "" {
		if token == "" {

		} else {

		}
	} else {
		if token == "" {

		} else {

		}
	}
	return files
}

func GetFileInfo(id string) (info map[string]string, err error) {
	info["fileID"] = id
	err = db.DB.View(func(tx *bolt.Tx) error {
		file := tx.Bucket(db.MyBucket).Bucket([]byte(id))
		if file == nil {
			return fmt.Errorf("file %s not found", id)
		}
		owner := file.Bucket([]byte("owner"))
		key, _ := owner.Cursor().First()
		info["owner"] = string(key)
		info["name"] = string(file.Get([]byte("name")))
		repo := file.Bucket([]byte("type"))
		if repo != nil {
			key, _ = repo.Cursor().First()
			info["repo"] = string(key)
		}
		if len(info["repo"]) == 0 {
			return fmt.Errorf("couldn't find repo for file %s", id)
		}
		info["version"] = string(file.Get([]byte("version")))
		info["tags"] = string(file.Get([]byte("tag")))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func MatchQuery(file, query map[string]string) bool {
	for key, value := range query {
		if file[key] != value {
			return false
		}
 	}
 	return true
}

func Search(query map[string]string) ([]SearchResult, error) {
	db.DB.View(func(tx *bolt.Tx) error {
		files := tx.Bucket(db.MyBucket)
		files.ForEach(func(k, v []byte) error {
			file, err := GetFileInfo(string(k))
			if err != nil {
				return err
			}
			if MatchQuery(file, query) {

			}
			return nil
		})
		return nil
	})
	return nil, nil
}