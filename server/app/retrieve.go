package app

import (
	"fmt"
	"time"
	"strings"
	"net/http"
	"github.com/boltdb/bolt"
	"github.com/blang/semver"
	"github.com/fatih/structs"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"github.com/mitchellh/mapstructure"
)

type SearchRequest struct {
	FileID    string `json:"FileID,omitempty"`    // files' UUID (or MD5)
	Owner     string `json:"Owner,omitempty"`     // files' owner username
	Name      string `json:"Name,omitempty"`      // files' name within CDN
	Repo      string `json:"Repo,omitempty"`      // files' repository - either "apt", "raw", or "template"
	Version   string `json:"Version,omitempty"`   // files' version
	Tags      string `json:"Tags,omitempty"`      // files' tags in format: "tag1,tag2,tag3"
	Token     string `json:"Token,omitempty"`     // user's token
	Verified  string `json:"Verified,omitempty"`  // flag for searching only among verified CDN users
	Operation string `json:"Operation,omitempty"` // operation type requested
}

type validateFunc func(SearchRequest) error

var (
	validators    = make(map[string]validateFunc)
	verifiedUsers = []string{"subutai", "jenkins", "docker", "travis", "appveyor", "devops"}
)

func InitValidators() {
	validators["info"] = ValidateInfo
	validators["list"] = ValidateList
}

func (request SearchRequest) ValidateRequest() error {
	validator := validators[request.Operation]
	return validator(request)
}

func CheckOwner(owner string) bool {
	exists := false
	db.DB.View(func(tx *bolt.Tx) error {
		exists = tx.Bucket(db.Users).Bucket([]byte(owner)) != nil
		return nil
	})
	return exists
}

func CheckToken(token string) bool {
	return db.TokenOwner(token) != ""
}

func In(item string, list []string) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}

func ValidateInfo(request SearchRequest) error {
	if request.FileID == "" && request.Name == "" {
		return fmt.Errorf("both fileID and name weren't given")
	}
	if request.FileID != "" && request.Name != "" && db.NameByHash(request.FileID) != request.Name {
		return fmt.Errorf("both fileID and name provided but they are not the same")
	}
	if !CheckOwner(request.Owner) {
		request.Owner = ""
	}
	if !CheckToken(request.Token) {
		request.Token = ""
	}
	if request.Verified == "true" && len(request.Owner) > 0 && !In(request.Owner, verifiedUsers) {
		return fmt.Errorf("both verified = true and owner given but owner is not a verified user")
	}
	return nil
}

func ValidateList(request SearchRequest) error {
	if !CheckOwner(request.Owner) {
		request.Owner = ""
	}
	if !CheckToken(request.Token) {
		request.Token = ""
	}
	if request.Verified == "true" && len(request.Owner) > 0 && !In(request.Owner, verifiedUsers) {
		return fmt.Errorf("both verified = true and owner given but owner is not a verified user")
	}
	return nil
}

// ParseRequest takes HTTP request and converts it into Request struct
func (request *SearchRequest) ParseRequest(httpRequest *http.Request) error {
	request.FileID    = httpRequest.URL.Query().Get("id")
	request.Owner     = httpRequest.URL.Query().Get("owner")
	request.Name      = httpRequest.URL.Query().Get("name")
	request.Repo      = strings.Split(httpRequest.RequestURI, "/")[3] // Splitting /kurjun/rest/repo/func into ["", "kurjun", "rest", "repo" (index: 3), "func?..."]
	request.Version   = httpRequest.URL.Query().Get("version")
	request.Tags      = httpRequest.URL.Query().Get("tags")
	request.Token     = strings.ToLower(httpRequest.URL.Query().Get("token"))
	request.Verified  = strings.ToLower(httpRequest.URL.Query().Get("verified"))
	request.Operation = strings.Split(strings.Split(httpRequest.RequestURI, "/")[4], "?")[0]
	return request.ValidateRequest()
}

// BuildQuery constructs the query out of the existing parameters in SearchRequest
func (request *SearchRequest) BuildQuery() (query map[string]string) {
	m := structs.Map(request)
	for k, v := range m {
		query[k] = v.(string)
		if query[k] == "latest" {
			query[k] = ""
		}
	}
	return
}

// SearchResult is a struct which return after search in db by parameters of SearchRequest
type SearchResult struct {
	FileID        string `json:"fileID,omitempty"`
	Owner         string `json:"owner,omitempty"`
	Name          string `json:"name,omitempty"`
	Filename      string `json:"filename,omitempty"`
	Repo          string `json:"type,omitempty"`
	Version       string `json:"version,omitempty"`
	Scope         string `json:"scope,omitempty"`
	Md5           string `json:"md5,omitempty"`
	Sha256        string `json:"sha256,omitempty"`
	Size          int    `json:"size,omitempty"`
	Tags          string `json:"tags,omitempty"`
	Date          string `json:"date,omitempty"`
	Timestamp     string `json:"timestamp,omitempty"`
	Description   string `json:"description,omitempty"`
	Architecture  string `json:"architecture,omitempty"`
	Parent        string `json:"parent,omitempty"`
	ParentVersion string `json:"parentVersion,omitempty"`
	ParentOwner   string `json:"parentOwner,omitempty"`
	PrefSize      string `json:"prefSize,omitempty"`
}

// BuildResult makes SearchResult struct from map
func BuildResult(info map[string]string) (result SearchResult) {
	mapstructure.Decode(info, &result)
	return
}

type filterFunc func(map[string]string, []SearchResult) []SearchResult

var (
	filters = make(map[string]filterFunc)
)

func InitFilters() {
	filters["info"] = FilterInfo
	filters["list"] = FilterList
}

func FilterInfo(query map[string]string, files []SearchResult) (result []SearchResult) {
	queryVersion, _ := semver.Make(query["version"])
	for _, file := range files {
		fileVersion, _ := semver.Make(file.Version)
		if query["version"] == "" {
			if fileVersion.GTE(queryVersion) {
				queryVersion = fileVersion
				result = []SearchResult{file}
			} else if fileVersion.EQ(queryVersion) && len(result) > 0 {
				resultDate, _ := time.Parse(time.RFC3339, result[0].Date)
				fileDate, _ := time.Parse(time.RFC3339, file.Date)
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
	return files
}

func Retrieve(request SearchRequest) []SearchResult {
	query := request.BuildQuery()
	files, err := Search(query)
	if err != nil {
		log.Error("retrieve couldn't search the query %+v", query)
		return nil
	}
	filter := filters[request.Operation]
	results := filter(query, files)
	return results
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
		info["date"] = string(file.Get([]byte("date")))
		if hash := tx.Bucket(db.MyBucket).Bucket([]byte(id)).Bucket([]byte("hash")); hash != nil {
			hash.ForEach(func(k, v []byte) error {
				if string(k) == "md5" {
					info["md5"] = string(v)
				}
				if string(k) == "sha256" {
					info["sha256"] = string(v)
				}
				return nil
			})
		}
		sz := file.Get([]byte("size"))
		if sz != nil {
			info["size"] = string(sz)
		} else {
			info["size"] = string(file.Get([]byte("Size")))
		}
		info["description"] = string(file.Get([]byte("Description")))
		arch := file.Get([]byte("Architecture"))
		if arch != nil {
			info["architecture"] = string(arch)
		} else {
			info["architecture"] = string(file.Get([]byte("arch")))
		}
		info["parent"] = string(file.Get([]byte("parent")))
		info["parentVersion"] = string(file.Get([]byte("parent-version")))
		info["parentOwner"] = string(file.Get([]byte("parent-owner")))
		info["prefSize"] = string(file.Get([]byte("prefsize")))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func MatchQuery(file, query map[string]string) bool {
	for key, value := range query {
		if key == "token" || key == "verified" {
			continue
		}
		if file[key] != value {
			return false
		}
	}
	if query["verified"] == "true" && !In(file["owner"], verifiedUsers) {
		return false
	}
	if !(query["verified"] == "true") && query["token"] != "" && !db.CheckShare(file["fileID"], db.TokenOwner(query["token"])) {
		return false
	}
	return true
}

// Search return list of files with parameters like query
func Search(query map[string]string) (list []SearchResult, err error) {
	var sr SearchResult
	db.DB.View(func(tx *bolt.Tx) error {
		files := tx.Bucket(db.MyBucket)
		files.ForEach(func(k, v []byte) error {
			file, err := GetFileInfo(string(k))
			if err != nil {
				return err
			}
			if MatchQuery(file, query) {
				sr = BuildResult(file)
				list = append(list, sr)
			}
			return nil
		})
		return nil
	})
	return list, nil
}
