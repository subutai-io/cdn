package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/boltdb/bolt"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/db"
)

type SearchRequest struct {
	fileID    string // files' UUID (or MD5)
	owner     string // files' owner username
	name      string // files' name within CDN
	repo      string // files' repository - either "apt", "raw", or "template"
	version   string // files' version
	tags      string // files' tags in format: "tag1,tag2,tag3"
	token     string // user's token
	verified  string // flag for searching only among verified CDN users
	operation string // operation type requested
}

// ParseRequest takes HTTP request and converts it into Request struct
func (request *SearchRequest) ParseRequest(httpRequest *http.Request) (err error) {
	request.fileID = httpRequest.URL.Query().Get("id")
	request.owner = httpRequest.URL.Query().Get("owner")
	request.name = httpRequest.URL.Query().Get("name")
	request.repo = strings.Split(httpRequest.RequestURI, "/")[3] // Splitting /kurjun/rest/repo/func into ["", "kurjun", "rest", "repo" (index: 3), "func?..."]
	request.version = httpRequest.URL.Query().Get("version")
	request.tags = httpRequest.URL.Query().Get("tags")
	request.token = httpRequest.URL.Query().Get("token")
	request.verified = httpRequest.URL.Query().Get("verified")
	request.operation = strings.Split(httpRequest.RequestURI, "/")[4][:4]
	return
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
	validator := validators[request.operation]
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
	if request.fileID == "" && request.name == "" {
		return fmt.Errorf("both fileID and name weren't given")
	}
	if request.fileID != "" && request.name != "" && db.NameByHash(request.fileID) != request.name {
		return fmt.Errorf("both fileID and name provided but they are not the same")
	}
	if !CheckOwner(request.owner) {
		request.owner = ""
	}
	if !CheckToken(request.token) {
		request.token = ""
	}
	if request.verified == "true" && len(request.owner) > 0 && !In(request.owner, verifiedUsers) {
		return fmt.Errorf("both verified = true and owner given but owner is not a verified user")
	}
	return nil
}

func ValidateList(request SearchRequest) error {
	if !CheckOwner(request.owner) {
		request.owner = ""
	}
	if !CheckToken(request.token) {
		request.token = ""
	}
	if request.verified == "true" && len(request.owner) > 0 && !In(request.owner, verifiedUsers) {
		return fmt.Errorf("both verified = true and owner given but owner is not a verified user")
	}
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
	if r.token != "" {
		query["token"] = r.token
	}
	if r.verified != "" {
		query["verified"] = r.verified
	}
	return
}

// SearchResult is a struct which return after search in db by parameters of SearchRequest
type SearchResult struct {
	fileID        string `json:"id,omitempty"`
	owner         string `json:"owner,omitempty"`
	name          string `json:"name,omitempty"`
	filename      string `json:"filename,omitempty"`
	repo          string `json:"type,omitempty"`
	version       string `json:"version,omitempty"`
	scope         string `json:"scope,omitempty"`
	md5           string `json:"md5,omitempty"`
	sha256        string `json:"sha256,omitempty"`
	size          int    `json:"size,omitempty"`
	tags          string `json:"tags,omitempty"`
	date          string `json:"date,omitempty"`
	timestamp     string `json:"timestamp,omitempty"`
	description   string `json:"description,omitempty"`
	architecture  string `json:"architecture,omitempty"`
	parent        string `json:"parent,omitempty"`
	parentVersion string `json:"parentVersion,omitempty"`
	parentOwner   string `json:"parentOwner,omitempty"`
	prefSize      string `json:"prefSize,omitempty"`
}

type filterFunc func(map[string]string, []SearchResult) []SearchResult

var (
	filters = make(map[string]filterFunc)
)

func InitFilters() {
	filters["info"] = FilterInfo
	filters["list"] = FilterList
}

// BuildResult makes SearchResult struct from map
func BuildResult(info map[string]string) (result SearchResult) {
	for k, v := range info {
		if k == "fileID" {
			result.fileID = v
		} else if k == "owner" {
			result.owner = v
		} else if k == "name" {
			result.name = v
		} else if k == "filename" {
			result.filename = v
		} else if k == "repo" {
			result.repo = v
		} else if k == "version" {
			result.version = v
		} else if k == "scope" {
			result.scope = v
		} else if k == "md5" {
			result.md5 = v
		} else if k == "sha256" {
			result.sha256 = v
		} else if k == "size" {
			size, _ := strconv.Atoi(v)
			result.size = size
		} else if k == "tags" {
			result.tags = v
		} else if k == "date" {
			result.date = v
		} else if k == "timestamp" {
			result.timestamp = v
		} else if k == "description" {
			result.description = v
		} else if k == "architecture" {
			result.architecture = v
		} else if k == "parent" {
			result.parent = v
		} else if k == "parentVersion" {
			result.parentVersion = v
		} else if k == "parentOwner" {
			result.parentOwner = v
		} else if k == "prefSize" {
			result.prefSize = v
		} else {
			log.Warn(fmt.Sprintf("Unrecognized key %s", k))
		}
	}
	return result
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
	if query["token"] != "" && !db.CheckShare(file["fileID"], db.TokenOwner(query["token"])) {
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
