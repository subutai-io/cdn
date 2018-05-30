package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/db"
)

type ValidateFunction func() error

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

	validators map[string]ValidateFunction
}

func (request *SearchRequest) InitValidators() {
	request.validators = make(map[string]ValidateFunction)
	request.validators["info"] = request.ValidateInfo
	request.validators["list"] = request.ValidateList
}

func (request *SearchRequest) ValidateRequest() error {
	validator := request.validators[request.Operation]
	return validator()
}

func (request *SearchRequest) ValidateInfo() error {
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

func (request *SearchRequest) ValidateList() error {
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
	request.FileID = httpRequest.URL.Query().Get("id")
	request.Owner = httpRequest.URL.Query().Get("owner")
	request.Name = httpRequest.URL.Query().Get("name")
	request.Repo = strings.Split(httpRequest.RequestURI, "/")[3] // Splitting /kurjun/rest/repo/func into ["", "kurjun", "rest", "repo" (index: 3), "func?..."]
	request.Version = httpRequest.URL.Query().Get("version")
	request.Tags = httpRequest.URL.Query().Get("tags")
	request.Token = strings.ToLower(httpRequest.URL.Query().Get("token"))
	request.Verified = strings.ToLower(httpRequest.URL.Query().Get("verified"))
	request.Operation = strings.Split(strings.Split(httpRequest.RequestURI, "/")[4], "?")[0]
	return request.ValidateRequest()
}

// BuildQuery constructs the query out of the existing parameters in SearchRequest
func (request *SearchRequest) BuildQuery() (query map[string]string) {
	query = make(map[string]string)
	m := structs.Map(request)
	for k, v := range m {
		if k == "validators" {
			continue
		}
		value := v.(string)
		if value == "" {
			continue
		}
		query[k] = v.(string)
		if query[k] == "latest" {
			query[k] = ""
		}
	}
	return
}

// Result represents all file's attributes
type Result struct {
	FileID        string `json:"FileID,omitempty"`
	Owner         string `json:"Owner,omitempty"`
	Name          string `json:"Name,omitempty"`
	Filename      string `json:"Filename,omitempty"`
	Repo          string `json:"Repo,omitempty"`
	Version       string `json:"Version,omitempty"`
	Scope         string `json:"Scope,omitempty"`
	Md5           string `json:"Md5,omitempty"`
	Sha256        string `json:"Sha256,omitempty"`
	Size          int64  `json:"Size,omitempty"`
	Tags          string `json:"Tags,omitempty"`
	Date          string `json:"Date,omitempty"`
	Timestamp     string `json:"Timestamp,omitempty"`
	Description   string `json:"Description,omitempty"`
	Architecture  string `json:"Architecture,omitempty"`
	Parent        string `json:"Parent,omitempty"`
	ParentVersion string `json:"ParentVersion,omitempty"`
	ParentOwner   string `json:"ParentOwner,omitempty"`
	PrefSize      string `json:"PrefSize,omitempty"`
}

func (result *Result) ConvertToOld() *OldResult {
	oldResult := new(OldResult)
	oldResult.FileID = result.FileID
	oldResult.Owner = []string{result.Owner}
	oldResult.Name = result.Name
	oldResult.Filename = result.Filename
	oldResult.Version = result.Version
	oldResult.Hash.Md5 = result.Md5
	oldResult.Hash.Sha256 = result.Sha256
	oldResult.Size = result.Size
	oldResult.Tags = strings.Split(result.Tags, ",")
	oldResult.Date = result.Date
	oldResult.Timestamp = result.Timestamp
	oldResult.Description = result.Description
	oldResult.Architecture = result.Architecture
	oldResult.Parent = result.Parent
	oldResult.ParentVersion = result.ParentVersion
	oldResult.ParentOwner = result.ParentOwner
	oldResult.PrefSize = result.PrefSize
	return oldResult
}

// BuildResult makes Result struct from map
func (result *Result) BuildResult(info map[string]string) {
	mapstructure.Decode(info, &result)
	return
}

type FilterFunction func(map[string]string, []*Result) []*Result

var (
	filters = make(map[string]FilterFunction)
)

func InitFilters() {
	log.Info("Initializing filters")
	filters["info"] = FilterInfo
	filters["list"] = FilterList
}

func FilterInfo(query map[string]string, files []*Result) (result []*Result) {
	queryVersion, _ := semver.Make(query["Version"])
	for _, file := range files {
		fileVersion, _ := semver.Make(file.Version)
		if query["Version"] == "" {
			if fileVersion.GTE(queryVersion) {
				queryVersion = fileVersion
				result = []*Result{file}
			} else if fileVersion.EQ(queryVersion) && len(result) > 0 {
				resultDate, _ := time.Parse(time.RFC3339, result[0].Date)
				fileDate, _ := time.Parse(time.RFC3339, file.Date)
				if resultDate.After(fileDate) {
					result = []*Result{file}
				}
			}
		} else {
			if queryVersion.EQ(fileVersion) {
				result = []*Result{file}
			}
		}
	}
	return
}

func FilterList(query map[string]string, files []*Result) (results []*Result) {
	return files
}

func (request *SearchRequest) Retrieve() []*Result {
	query := request.BuildQuery()
	log.Info("Query: ", query)
	files, err := Search(query)
	log.Info("Files: ", files)
	if err != nil {
		log.Error("retrieve couldn't search the query %+v", query)
		return nil
	}
	filter := filters[request.Operation]
	results := filter(query, files)
	return results
}

func GetFileInfo(id string) (info map[string]string, err error) {
	info = make(map[string]string)
	info["FileID"] = id
	err = db.DB.View(func(tx *bolt.Tx) error {
		file := tx.Bucket(db.MyBucket).Bucket([]byte(id))
		if file == nil {
			return fmt.Errorf("file %s not found", id)
		}
		owner := file.Bucket([]byte("owner"))
		key, _ := owner.Cursor().First()
		info["Owner"] = string(key)
		info["Name"] = string(file.Get([]byte("name")))
		repo := file.Bucket([]byte("type"))
		if repo != nil {
			key, _ = repo.Cursor().First()
			info["Repo"] = string(key)
		}
		if len(info["Repo"]) == 0 {
			return fmt.Errorf("couldn't find repo for file %s", id)
		}
		info["Version"] = string(file.Get([]byte("version")))
		info["Tags"] = string(file.Get([]byte("tag")))
		info["Date"] = string(file.Get([]byte("date")))
		date, _ := time.Parse(time.RFC3339Nano, info["Date"])
		info["Timestamp"] = strconv.FormatInt(date.Unix(), 10)
		if hash := tx.Bucket(db.MyBucket).Bucket([]byte(id)).Bucket([]byte("hash")); hash != nil {
			hash.ForEach(func(k, v []byte) error {
				if string(k) == "md5" {
					info["Md5"] = string(v)
				}
				if string(k) == "sha256" {
					info["Sha256"] = string(v)
				}
				return nil
			})
		}
		sz := file.Get([]byte("size"))
		if sz != nil {
			info["Size"] = string(sz)
		} else {
			info["Size"] = string(file.Get([]byte("Size")))
		}
		info["Description"] = string(file.Get([]byte("Description")))
		arch := file.Get([]byte("Architecture"))
		if arch != nil {
			info["Architecture"] = string(arch)
		} else {
			info["Architecture"] = string(file.Get([]byte("arch")))
		}
		info["Parent"] = string(file.Get([]byte("parent")))
		info["ParentVersion"] = string(file.Get([]byte("parent-version")))
		info["ParentOwner"] = string(file.Get([]byte("parent-owner")))
		info["PrefSize"] = string(file.Get([]byte("prefsize")))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func MatchQuery(file, query map[string]string) bool {
	for key, value := range query {
		if key == "Token" || key == "Verified" || key == "Operation" {
			continue
		}
		if file[key] != value {
			log.Info(fmt.Sprintf("Key: %v, value: %v, file: %v", key, value, file[key]))
			return false
		}
	}
	if (query["Token"] == "" && !db.IsPublic(file["FileID"])) ||
		(query["Token"] != "" && !db.CheckShare(file["FileID"], db.TokenOwner(query["Token"]))) {
		return false
	}
	if query["Verified"] == "true" && !In(file["Owner"], verifiedUsers) {
		return false
	}
	if !(query["Verified"] == "true") && query["Token"] != "" && !db.CheckShare(file["FileID"], db.TokenOwner(query["Token"])) {
		return false
	}
	return true
}

// Search return list of files with parameters like query
func Search(query map[string]string) (list []*Result, err error) {
	db.DB.View(func(tx *bolt.Tx) error {
		files := tx.Bucket(db.MyBucket)
		files.ForEach(func(k, v []byte) error {
			file, err := GetFileInfo(string(k))
			if err != nil {
				log.Info("error")
				return err
			}
			//log.Info("File: ", file)
			if MatchQuery(file, query) {
				log.Info("File and query are equal")
				sr := new(Result)
				sr.BuildResult(file)
				list = append(list, sr)
			}
			return nil
		})
		return nil
	})
	return list, nil
}
