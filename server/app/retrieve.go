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

	admin      string
	validators map[string]ValidateFunction
}

func (request *SearchRequest) InitValidators() {
	request.validators = make(map[string]ValidateFunction)
	request.validators["info"] = request.ValidateInfo
	request.validators["list"] = request.ValidateList
}

func (request *SearchRequest) ValidateRequest() error {
	if !In(request.Operation, []string{"info", "list"}) {
		return fmt.Errorf("incorrect SearchRequest: specify request.Operation type - either \"info\" or \"list\"")
	}
	validator := request.validators[request.Operation]
	return validator()
}

func (request *SearchRequest) ValidateInfo() error {
	if request.FileID == "" && request.Name == "" {
		return fmt.Errorf("both fileID and name weren't given")
	}
	if request.FileID != "" && request.Name != "" && DB.NameByHash(request.FileID) != request.Name {
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
	rest := httpRequest.URL.EscapedPath()
	request.Repo = strings.Split(rest, "/")[3] // Splitting /kurjun/rest/repo/func into ["", "kurjun", "rest", "repo" (index: 3), "func?..."]
	request.Version = httpRequest.URL.Query().Get("version")
	request.Tags = httpRequest.URL.Query().Get("tags")
	request.Token = strings.ToLower(httpRequest.URL.Query().Get("token"))
	request.Verified = strings.ToLower(httpRequest.URL.Query().Get("verified"))
	request.Operation = strings.Split(strings.Split(rest, "/")[4], "?")[0]
	if DB.TokenOwner(request.Token) == "subutai" {
		request.admin = "true"
	}
	return request.ValidateRequest()
}

// BuildQuery constructs the query out of the existing parameters in SearchRequest
func (request *SearchRequest) BuildQuery() (query map[string]string) {
	query = make(map[string]string)
	m := structs.Map(request)
	for k, v := range m {
		log.Debug(fmt.Sprintf("Building query - key: %+v, value: %+v", k, v))
		value := v.(string)
		if value == "" {
			continue
		}
		query[k] = v.(string)
		if query[k] == "latest" {
			query[k] = ""
		}
	}
	if request.admin != "" {
		query["admin"] = request.admin
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
	log.Info(fmt.Sprintf("Converting result %+v to oldResult", result))
	oldResult := &OldResult{
		FileID:   result.FileID,
		Owner:    []string{result.Owner},
		Name:     result.Name,
		Filename: result.Filename,
		Version:  result.Version,
		Hash: Hashes{
			Md5:    result.Md5,
			Sha256: result.Sha256,
		},
		Size:          result.Size,
		Tags:          strings.Split(result.Tags, ","),
		Date:          result.Date,
		Timestamp:     result.Timestamp,
		Description:   result.Description,
		Architecture:  result.Architecture,
		Parent:        result.Parent,
		ParentVersion: result.ParentVersion,
		ParentOwner:   result.ParentOwner,
		PrefSize:      result.PrefSize,
	}
	return oldResult
}

// BuildResult makes Result struct from map
func (result *Result) BuildResult(info map[string]string) {
	mapstructure.Decode(info, &result)
	return
}

func GetResultByFileID(fileID string) (result *Result) {
	info, err := GetFileInfo(fileID)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't build result by fileID: %v", err))
	} else {
		result = new(Result)
		result.BuildResult(info)
	}
	return
}

type FilterFunction func(map[string]string, []*Result) []*Result

var (
	filters map[string]FilterFunction
)

func InitFilters() {
	filters = make(map[string]FilterFunction)
	filters["info"] = FilterInfo
	filters["list"] = FilterList
	log.Info("Initialization of filters finished")
}

func FilterInfo(query map[string]string, files []*Result) (result []*Result) {
	log.Info(fmt.Sprintf("FilterInfo: %+v, files: %+v", query, files))
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
	log.Info(fmt.Sprintf("FilterList: %+v, files: %+v", query, files))
	return files
}

func (request *SearchRequest) Retrieve() (results []*Result) {
	query := request.BuildQuery()
	log.Info("Query: ", query)
	files, err := Search(query)
	log.Info("Files: ", files)
	if err != nil {
		log.Error("retrieve couldn't search the query %+v", query)
		return nil
	}
	if !In(request.Operation, []string{"info", "list"}) {
		return
	}
	filter := filters[request.Operation]
	log.Info(fmt.Sprintf("request.Operation = %s", request.Operation))
	log.Info(fmt.Sprintf("query = %+v", query))
	log.Info(fmt.Sprintf("files = %+v", files))
	results = filter(query, files)
	log.Info(fmt.Sprintf("results = %+v", results))
	return
}

func MatchQuery(file, query map[string]string) bool {
	log.Info(fmt.Sprintf("\nfile: %+v\n\nquery: %+v", file, query))
	for key, value := range query {
		if key == "Token" || key == "Verified" || key == "Operation" || key == "admin" {
			continue
		}
		if file[key] != value {
			log.Info(fmt.Sprintf("Key: %v, value: %v, file: %v", key, value, file[key]))
			return false
		}
	}
	fileID := file["FileID"]
	token := file["Token"]
	verified := file["Verified"]
	shared := DB.IsPublic(fileID)
	if token != "" && DB.CheckShare(fileID, DB.TokenOwner(token)) {
		shared = true
	}
	if verified == "true" && !In(file["Owner"], verifiedUsers) {
		log.Warn("file unavailable because user is not verified")
		return false
	}
	if query["admin"] != "true" {
		log.Info("Verdict: %+v", shared)
		return shared
	} else {
		return true
	}
}
