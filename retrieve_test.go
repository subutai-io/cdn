package app

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/subutai-io/agent/log"
)

func TestValidateRequest(t *testing.T) {
	request := new(SearchRequest)
	request.Operation = "operation"
	if err := request.ValidateRequest(); err == nil {
		t.Error(err)
	}
}

func TestSearchRequest_ValidateInfo(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type fields struct {
		FileID   string
		Owner    string
		Name     string
		Token    string
		Verified string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "t1", fields: fields{}},
		{name: "t2", fields: fields{FileID: "id2", Name: "name2"}},
		{name: "t3", fields: fields{FileID: "id3", Owner: "subutai", Name: "name3"}},
		{name: "t4", fields: fields{FileID: "id4", Owner: "NotExistingOwner", Name: "name4"}},
		{name: "t5", fields: fields{FileID: "id5", Owner: "Owner", Name: "name5", Verified: "true"}},
		{name: "t6", fields: fields{FileID: "id6", Owner: "subutai", Name: "name6"}},
		{name: "t7", fields: fields{FileID: "id7", Owner: "lorem", Name: "name7", Verified: "true"}},
	}
	for _, tt := range tests {
		errored := false
		request := &SearchRequest{
			FileID:   tt.fields.FileID,
			Owner:    tt.fields.Owner,
			Name:     tt.fields.Name,
			Token:    tt.fields.Token,
			Verified: tt.fields.Verified,
		}
		if tt.name == "t1" {
			if err := request.ValidateInfo(); err == nil {
				errored = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		} else if tt.name == "t2" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name, tt.fields.Owner, "raw")
			if err := request.ValidateInfo(); err != nil {
				errored = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		} else if tt.name == "t3" {
			PrepareUsersAndTokens()
			request.ValidateInfo()
			if request.Owner != tt.fields.Owner {
				errored = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Owner is different. Wait: %v, got: %v", tt.name, tt.fields.Name, request.Owner)
			}
		} else if tt.name == "t4" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name, tt.fields.Owner, "raw")
			request.ValidateInfo()
			if request.Owner != "" {
				t.Errorf("%q. SearchRequest.ValidateInfo. Owner is not exist and must be empty. Got: %v", tt.name, request.Owner)
			}
		} else if tt.name == "t6" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name, tt.fields.Owner, "raw")
			if err := request.ValidateInfo(); err != nil {
				errored = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		} else if tt.name == "t7" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name, tt.fields.Owner, "raw")
			if err := request.ValidateInfo(); err == nil {
				errored = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		}
		if errored {
			break
		}
	}
}

func TestSearchRequest_ValidateList(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type fields struct {
		FileID     string
		Owner      string
		Name       string
		Repo       string
		Version    string
		Tags       string
		Token      string
		Verified   string
		Operation  string
		admin      string
		validators map[string]ValidateFunction
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "t1", fields: fields{Owner: "subutai", Token: Subutai.Token}},
		{name: "t2", fields: fields{Owner: "owner"}},
		{name: "t3", fields: fields{Owner: "lorem", Token: Lorem.Token, Verified: "true"}},
	}
	for _, tt := range tests {
		request := &SearchRequest{
			FileID:     tt.fields.FileID,
			Owner:      tt.fields.Owner,
			Name:       tt.fields.Name,
			Repo:       tt.fields.Repo,
			Version:    tt.fields.Version,
			Tags:       tt.fields.Tags,
			Token:      tt.fields.Token,
			Verified:   tt.fields.Verified,
			Operation:  tt.fields.Operation,
			admin:      tt.fields.admin,
			validators: tt.fields.validators,
		}
		if tt.name == "t1" {
			if err := request.ValidateList(); err != nil {
				t.Errorf("%q. SearchRequest.ValidateList() error = %v", tt.name, err)
			}
		}
		if tt.name == "t2" {
			request.ValidateList()
			if request.Owner != "" && request.Token == "" {
				t.Errorf("%q. SearchRequest.ValidateList(). Owner must be empty. Got: %v", tt.name, tt.fields.Owner)
			}
		}
		if tt.name == "t3" {
			if err := request.ValidateList(); err == nil {
				t.Errorf("%q. SearchRequest.ValidateList() error = %v", tt.name, err)
			}
		}
	}
}

func TestSearchRequest_ParseRequest(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	TearDown()
	type fields struct {
		FileID     string
		Owner      string
		Name       string
		Repo       string
		Version    string
		Tags       string
		Token      string
		Verified   string
		Operation  string
		validators map[string]ValidateFunction
	}
	type args struct {
		httpRequest *http.Request
	}

	request1, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/raw/info?id=id1&owner=owner1&name=name1&version=version1&tags=tag1&token=token1&verified=false", nil)
	request2, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/apt/info?id=id2&owner=owner2&name=name2&version=version2&tags=tag2&token=token2&verified=false", nil)
	request3, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/template/info?id=id3&owner=owner3&name=name3", nil)
	request4, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/template/info?id=id4&owner=subutai&name=name4&&verified=true&&token=15a5237ee5314282e52156cfad72e86b53ef0ad47baecc31233dbb1c06f4327c", nil)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "name1", Repo: "raw", Version: "version1", Tags: "tag1",
			Token: "token1", Verified: "false", Operation: "info"}, args: args{httpRequest: request1}},
		{name: "t2", fields: fields{FileID: "id2", Owner: "owner2", Name: "name2", Repo: "apt", Version: "version2", Tags: "tag2",
			Token: "token2", Verified: "false", Operation: "info"}, args: args{httpRequest: request2}},
		{name: "t3", fields: fields{FileID: "id3", Owner: "owner3", Name: "name3", Repo: "", Version: "", Tags: "",
			Token: "", Verified: "", Operation: ""}, args: args{httpRequest: request3}},
		{name: "t4", fields: fields{FileID: "id4", Owner: "subutai", Name: "name4", Repo: "", Version: "", Tags: "",
			Token: Subutai.Token, Verified: "true", Operation: ""}, args: args{httpRequest: request4}},
	}
	for _, tt := range tests {
		SRequest := &SearchRequest{
			FileID:     tt.fields.FileID,
			Owner:      tt.fields.Owner,
			Name:       tt.fields.Name,
			Repo:       tt.fields.Repo,
			Version:    tt.fields.Version,
			Tags:       tt.fields.Tags,
			Token:      tt.fields.Token,
			Verified:   tt.fields.Verified,
			Operation:  tt.fields.Operation,
			validators: tt.fields.validators,
		}
		SRequest.InitValidators()
		SRequest.ParseRequest(tt.args.httpRequest)
		if SRequest.FileID != tt.fields.FileID &&
			SRequest.Owner != tt.fields.Owner &&
			SRequest.Name != tt.fields.Repo &&
			SRequest.Repo != tt.fields.Repo &&
			SRequest.Version != tt.fields.Version &&
			SRequest.Tags != tt.fields.Tags &&
			SRequest.Token != tt.fields.Token &&
			SRequest.Verified != tt.fields.Verified &&
			SRequest.Operation != tt.fields.Operation {
			t.Errorf("%q. Error. Wait %v, got %v", tt.name, tt.fields, SRequest)
		}
		if tt.name == "t4" {
			if SRequest.admin != "true" {
				t.Error("admin must be true")
			}
		}
	}
}

func TestSearchRequest_BuildQuery(t *testing.T) {
	Integration = 0
	type fields struct {
		FileID     string
		Owner      string
		Name       string
		Repo       string
		Version    string
		Tags       string
		Token      string
		Verified   string
		Operation  string
		admin      string
		validators map[string]ValidateFunction
	}

	wantQuery1 := map[string]string{"FileID": "id1", "Owner": "owner1", "Name": "name1", "Repo": "repo1", "Version": "", "Tags": "tag1", "Verified": "true", "Operation": "info", "admin": "true"}

	tests := []struct {
		name      string
		fields    fields
		wantQuery map[string]string
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "name1", Repo: "repo1", Version: "latest", Tags: "tag1", Verified: "true", Operation: "info", admin: "true"}, wantQuery: wantQuery1},
	}
	for _, tt := range tests {
		request := &SearchRequest{
			FileID:     tt.fields.FileID,
			Owner:      tt.fields.Owner,
			Name:       tt.fields.Name,
			Repo:       tt.fields.Repo,
			Version:    tt.fields.Version,
			Tags:       tt.fields.Tags,
			Token:      tt.fields.Token,
			Verified:   tt.fields.Verified,
			Operation:  tt.fields.Operation,
			admin:      tt.fields.admin,
			validators: tt.fields.validators,
		}
		if gotQuery := request.BuildQuery(); !reflect.DeepEqual(gotQuery, tt.wantQuery) {
			t.Errorf("%q. SearchRequest.BuildQuery() = %v, want %v", tt.name, gotQuery, tt.wantQuery)
		}
	}
}

func TestResult_ConvertToOld(t *testing.T) {
	Integration = 0
	type fields struct {
		FileID        string
		Owner         string
		Name          string
		Filename      string
		Repo          string
		Version       string
		Scope         string
		Md5           string
		Sha256        string
		Size          int64
		Tags          string
		Date          string
		Timestamp     string
		Description   string
		Architecture  string
		Parent        string
		ParentVersion string
		ParentOwner   string
		PrefSize      string
	}

	var owners, tags []string
	owners = append(owners, "owner1")
	tags = append(tags, "tag1")
	oldRes := new(OldResult)
	oldRes.FileID = "id1"
	oldRes.Owner = owners
	oldRes.Name = "name1"
	oldRes.Version = "v1"
	oldRes.Tags = tags
	oldRes.Description = "description"
	oldRes.Architecture = "amd64"
	oldRes.Hash.Md5 = "md5"
	oldRes.Hash.Sha256 = "sha256"
	tests := []struct {
		name   string
		fields fields
		want   *OldResult
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "name1", Md5: "md5", Sha256: "sha256", Version: "v1", Scope: "scope1", Tags: "tag1", Description: "description", Architecture: "amd64"}, want: oldRes},
	}
	for _, tt := range tests {
		result := &Result{
			FileID:        tt.fields.FileID,
			Owner:         tt.fields.Owner,
			Name:          tt.fields.Name,
			Filename:      tt.fields.Filename,
			Repo:          tt.fields.Repo,
			Version:       tt.fields.Version,
			Scope:         tt.fields.Scope,
			Md5:           tt.fields.Md5,
			Sha256:        tt.fields.Sha256,
			Size:          tt.fields.Size,
			Tags:          tt.fields.Tags,
			Date:          tt.fields.Date,
			Timestamp:     tt.fields.Timestamp,
			Description:   tt.fields.Description,
			Architecture:  tt.fields.Architecture,
			Parent:        tt.fields.Parent,
			ParentVersion: tt.fields.ParentVersion,
			ParentOwner:   tt.fields.ParentOwner,
			PrefSize:      tt.fields.PrefSize,
		}
		if got := result.ConvertToOld(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Result.ConvertToOld() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestResult_BuildResult(t *testing.T) {
	Integration = 0
	type fields struct {
		FileID        string
		Owner         string
		Name          string
		Filename      string
		Repo          string
		Version       string
		Scope         string
		Md5           string
		Sha256        string
		Size          int64
		Tags          string
		Date          string
		Timestamp     string
		Description   string
		Architecture  string
		Parent        string
		ParentVersion string
		ParentOwner   string
		PrefSize      string
	}
	info1 := map[string]string{"FileID": "id1"}
	info2 := map[string]string{"FileID": "id1", "Md5": "Md5", "Description": "Description"}
	tests := []struct {
		name   string
		fields fields
		args   map[string]string
	}{
		{name: "t1", fields: fields{FileID: "id1"}, args: info1},
		{name: "t1", fields: fields{FileID: "id1", Md5: "Md5", Description: "Description"}, args: info2},
	}
	for _, tt := range tests {
		var result Result
		result.BuildResult(tt.args)
		if result.FileID != tt.fields.FileID &&
			result.Owner != tt.fields.Owner && result.Name != tt.fields.Name &&
			result.Filename != tt.fields.Filename && result.Repo != tt.fields.Repo &&
			result.Version != tt.fields.Version && result.Scope != tt.fields.Scope &&
			result.Md5 != tt.fields.Md5 && result.Sha256 != tt.fields.Sha256 &&
			result.Tags != tt.fields.Tags && result.Date != tt.fields.Date &&
			result.Timestamp != tt.fields.Timestamp && result.Description != tt.fields.Description &&
			result.Architecture != tt.fields.Architecture && result.Parent != tt.fields.Parent &&
			result.ParentVersion != tt.fields.ParentVersion && result.ParentOwner != tt.fields.ParentOwner &&
			result.PrefSize != tt.fields.PrefSize {
			t.Errorf("%q. Result.BuildResult() = %v, want %v", tt.name, result, tt.fields)
		}
	}
}

func TestGetResultByFileID(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	result := new(Result)
	result.FileID = "id1"
	result.Filename = "file1"
	result.Name = "file1"
	result.Owner = "subutai"
	result.Repo = "raw"
//	WriteDB(result)
	type args struct {
		fileID string
	}
	tests := []struct {
		name       string
		args       args
		wantResult *Result
	}{
		{name: "t1", args: args{fileID: "id1"}, wantResult: result},
		{name: "t2", args: args{fileID: "id2"}, wantResult: result},
	}
	for _, tt := range tests {
		if tt.name == "t1" {
			gotResult := GetResultByFileID(tt.args.fileID)
			if gotResult.FileID != tt.wantResult.FileID &&
				gotResult.Filename != tt.wantResult.Filename &&
				gotResult.Name != tt.wantResult.Name &&
				gotResult.Owner != tt.wantResult.Owner &&
				gotResult.Repo != tt.wantResult.Repo {
				t.Errorf("%q. GetResultByFileID() = %v, want %v", tt.name, gotResult, tt.wantResult)
			}
		} else if tt.name == "t2" {
			if result := GetResultByFileID(tt.args.fileID); result != nil {
				t.Errorf("%q. GetResultByFileID() ", tt.name)
			}
		}
	}
}

func TestFilterInfo(t *testing.T) {
	Integration = 0
	SetUp()
	type args struct {
		query map[string]string
		files []*Result
	}
	result1 := new(Result)
	result1.Name = "file1"
	result1.Repo = "raw"
	result1.Tags = "tag1"
	result1.FileID = "id1"
	result1.Owner = "subutai"
	result1.Version = "1"
	result1.Filename = "file1"
	result1.Date = " 2018-06-09T07:52:15.208754096+06:00"

	result2 := new(Result)
	result2.Name = "file2"
	result2.Repo = "raw"
	result2.Tags = "tag1"
	result2.FileID = "id2"
	result2.Owner = "subutai"
	result2.Version = "0"
	result2.Filename = "file1"
	result2.Date = " 2018-06-09T07:53:15.208754096+06:00"
	var results []*Result
	results = append(results, result1)
	results = append(results, result2)

	var wantResult1 []*Result
	wantResult1 = append(wantResult1, result2)
	query := map[string]string{"FileID": "id1", "Name": "file1", "Filename": "file1", "Repo": "raw", "Tags": "tag1", "Owner": "subutai", "Version": "1", "Token": Subutai.Token}
	query2 := map[string]string{"FileID": "id1", "Name": "file1", "Filename": "file1", "Repo": "raw", "Tags": "tag1", "Owner": "subutai", "Version": "0", "Token": Subutai.Token}
	tests := []struct {
		name       string
		args       args
		wantResult []*Result
	}{
		{name: "t1", args: args{query: query, files: results}, wantResult: wantResult1},
		{name: "t2", args: args{query: query2, files: results}, wantResult: wantResult1},
	}
	for _, tt := range tests {
		errored := false
		if tt.name == "t1" {
			if gotResult := FilterInfo(tt.args.query, tt.args.files); !reflect.DeepEqual(gotResult, tt.wantResult) {
				errored = true
				t.Errorf("%q. FilterInfo() = %v, want %v", tt.name, gotResult, tt.wantResult)
			}
		} else if tt.name == "t2" {
			if gotResult := FilterInfo(tt.args.query, tt.args.files); !reflect.DeepEqual(gotResult, tt.wantResult) {
				errored = true
				t.Errorf("%q. FilterInfo() = %v, want %v", tt.name, gotResult, tt.wantResult)
			}
		}
		if errored {
			break
		}
	}
	TearDown()
}

func TestFilterList(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	result1 := new(Result)
	result1.Name = "file1"
	result1.Repo = "raw"
	result1.Tags = "tag1"
	result1.FileID = "id1"
	result1.Owner = "subutai"
	result1.Version = "1"
	result1.Filename = "file1"
	result1.Date = " 2018-06-09T07:52:15.208754096+06:00"

	result2 := new(Result)
	result2.Name = "file2"
	result2.Repo = "raw"
	result2.Tags = "tag1"
	result2.FileID = "id2"
	result2.Owner = "subutai"
	result2.Version = "0"
	result2.Filename = "file1"
	result2.Date = " 2018-06-09T07:53:15.208754096+06:00"
	var results []*Result
	results = append(results, result1)
	results = append(results, result2)

	var wantResults []*Result
	wantResults = append(wantResults, result1)
	wantResults = append(wantResults, result2)

	type args struct {
		query map[string]string
		files []*Result
	}
	query := map[string]string{"FileID": "id1", "Name": "file1", "Filename": "file1", "Repo": "raw", "Tags": "tag1", "Owner": "subutai", "Version": "1", "Token": Subutai.Token}
	tests := []struct {
		name        string
		args        args
		wantResults []*Result
	}{
		{name: "t1", args: args{query: query, files: results}, wantResults: wantResults},
	}
	for _, tt := range tests {
		if gotResults := FilterList(tt.args.query, tt.args.files); !reflect.DeepEqual(gotResults, tt.wantResults) {
			t.Errorf("%q. FilterList() = %v, want %v", tt.name, gotResults, tt.wantResults)
		}
	}
}

func TestSearchRequest_Retrieve(t *testing.T) {
	Integration = 0
	SetUp()
	type fields struct {
		FileID     string
		Owner      string
		Name       string
		Repo       string
		Version    string
		Tags       string
		Token      string
		Verified   string
		Operation  string
		validators map[string]ValidateFunction
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "t1", fields: fields{FileID: "id", Name: "notexistedfile", Owner: "subutai", Repo: "raw", Operation: "list"}},
		{name: "t2", fields: fields{FileID: "id2", Repo: "raw", Operation: "info"}},
	}
	PrepareUsersAndTokens()
	for _, tt := range tests {
		errored := false
		request := &SearchRequest{
			FileID:     tt.fields.FileID,
			Owner:      tt.fields.Owner,
			Name:       tt.fields.Name,
			Repo:       tt.fields.Repo,
			Version:    tt.fields.Version,
			Tags:       tt.fields.Tags,
			Token:      tt.fields.Token,
			Verified:   tt.fields.Verified,
			Operation:  tt.fields.Operation,
			validators: tt.fields.validators,
		}
		if tt.name == "t1" {
			if results := request.Retrieve(); results != nil {
				errored = true
				t.Errorf("%q. SearchRequest.Retrieve(). File is not exist", tt.name)
			}
		}
		if tt.name == "t2" {
			WriteFileInDB(tt.fields.FileID, "file1", tt.fields.Owner, tt.fields.Repo)
			results := request.Retrieve()
			log.Info(fmt.Sprintf("%v. Results: %v", tt.name, results))
			for _, result := range results {
				if result.FileID != tt.fields.FileID && len(results) != 1 {
					t.Errorf("%q. SearchRequest.Retrieve()", tt.name)
				}
			}
		}
		if errored {
			break
		}
	}
	TearDown()
}

func TestGetFileInfo(t *testing.T) {
	Integration = 0
	SetUp()
	type args struct {
		fileID string
	}

	wantInfo1 := map[string]string{"FileID": "id1", "Filename": "file", "Owner": "subutai", "Repo": "raw"}
	wantInfo2 := map[string]string{}

	tests := []struct {
		name     string
		args     args
		wantInfo map[string]string
	}{
		{name: "t1", args: args{fileID: "id1"}, wantInfo: wantInfo1},
		{name: "t2", args: args{fileID: "id2"}, wantInfo: wantInfo2},
	}
	PrepareUsersAndTokens()
	for _, tt := range tests {
		errored := false
		if tt.name == "t1" {
			WriteFileInDB(tt.args.fileID, "file", "subutai", "raw")
			gotInfo, err := GetFileInfo(tt.args.fileID)
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) && err != nil {
				errored = true
				t.Errorf("%q. GetFileInfo() = %v, want %v", tt.name, gotInfo, tt.wantInfo)
			}
		} else if tt.name == "t2" {
			WriteFileWithoutRepo(tt.args.fileID, "file2", "subutai")
			gotInfo, err := GetFileInfo(tt.args.fileID)
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) && err == nil {
				errored = true
				t.Errorf("%q. GetFileInfo() = %v, want %v", tt.name, gotInfo, tt.wantInfo)
			}
		}
		if errored {
			break
		}
	}
	TearDown()
}

func TestMatchQuery(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type args struct {
		file  map[string]string
		query map[string]string
	}
	file1 := map[string]string{"FileID": "id1", "Filename": "file1", "Owner": "subutai", "Token": Subutai.Token, "Repo": "raw"}
	query1 := map[string]string{"FileID": "id1", "Filename": "file1", "Owner": "subutai", "Token": Subutai.Token, "Repo": "raw"}
	file2 := map[string]string{"FileID": "id2", "Filename": "file2", "Owner": "lorem", "Verified": "true", "Repo": "raw"}
	query2 := map[string]string{"FileID": "id2", "Filename": "file2", "Owner": "lorem", "Verified": "true", "Repo": "raw"}
	tests := []struct {
		name string
		args args
	}{
		{name: "t1", args: args{file: file1, query: query1}},
		{name: "t2", args: args{file: file2, query: query2}},
	}
	for _, tt := range tests {
		if tt.name == "t1" {
			WriteFileInDB(tt.args.file["FileID"], tt.args.file["Filename"], tt.args.file["Owner"], tt.args.file["Repo"])
			if got := MatchQuery(tt.args.file, tt.args.query); !got {
				t.Errorf("%q. MatchQuery()", tt.name)
			}
		} else if tt.name == "t2" {
			WriteFileInDB(tt.args.file["FileID"], tt.args.file["Filename"], tt.args.file["Owner"], tt.args.file["Repo"])
			if got := MatchQuery(tt.args.file, tt.args.query); got {
				t.Errorf("%q. MatchQuery()", tt.name)
			}
		}
	}
}

func TestSearch(t *testing.T) {
	Integration = 0
	SetUp()
	type args struct {
		query map[string]string
	}
	PrepareUsersAndTokens()
	result1 := new(Result)
	result1.Name = "file1"
	result1.Repo = "raw"
	result1.Tags = "tag1"
	result1.FileID = "id1"
	result1.Owner = "subutai"
	result1.Version = "v1"
	result1.Filename = "file1"
	var results []*Result
	results = append(results, result1)
	query := map[string]string{"FileID": "id1", "Name": "file1", "Filename": "file1", "Repo": "raw", "Tags": "tag1", "Owner": "subutai", "Version": "v1", "Token": Subutai.Token}
	tests := []struct {
		name     string
		args     args
		wantList []*Result
	}{
		{name: "t1", args: args{query: query}, wantList: results},
	}
	for _, tt := range tests {
		errorred := false
		WriteFileInDB("FileID", tt.args.query["Name"], "subutai", tt.args.query["Repo"])
		WriteDB(result1)
		gotList, err := Search(tt.args.query)
		if err != nil {
			errorred = true
			t.Errorf("%q. Search() error = %v", tt.name, err)
			continue
		}
		if len(gotList) == 0 {
			log.Info("Got list is empty")
		}
		if len(gotList) != 0 {
			if gotList[0].FileID != tt.wantList[0].FileID &&
				gotList[0].Owner != tt.wantList[0].Owner &&
				gotList[0].Name != tt.wantList[0].Name &&
				gotList[0].Filename != tt.wantList[0].Filename &&
				gotList[0].Repo != tt.wantList[0].Repo &&
				gotList[0].Version != tt.wantList[0].Version &&
				gotList[0].Scope != tt.wantList[0].Scope &&
				gotList[0].Md5 != tt.wantList[0].Md5 &&
				gotList[0].Sha256 != tt.wantList[0].Sha256 &&
				gotList[0].Tags != tt.wantList[0].Tags &&
				gotList[0].Description != tt.wantList[0].Description &&
				gotList[0].Architecture != tt.wantList[0].Architecture &&
				gotList[0].Parent != tt.wantList[0].Parent &&
				gotList[0].ParentVersion != tt.wantList[0].ParentVersion &&
				gotList[0].ParentOwner != tt.wantList[0].ParentOwner &&
				gotList[0].PrefSize != tt.wantList[0].PrefSize {
				t.Error("Error in Search. Results is not equal")
			}
		}
		if errorred {
			break
		}
	}
	TearDown()
}
