package app

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"

	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/db"
)

func WriteOwnerInDB(owner string) {
	db.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(db.Users); b != nil {
			b.CreateBucket([]byte(owner))
			log.Info("Created owner ", owner)
		}
		return nil
	})
}

func WriteFileInDB(fileID, fileName, owner, repo string) {
	db.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(db.MyBucket); b != nil {
			if c, _ := b.CreateBucket([]byte(fileID)); c != nil {
				c.Put([]byte("name"), []byte(fileName))
				if d, _ := c.CreateBucket([]byte("owner")); d != nil {
					d.Put([]byte(owner), []byte("w"))
				}
				if d, _ := c.CreateBucket([]byte("type")); d != nil {
					d.Put([]byte(repo), []byte("w"))
				}
				c.CreateBucket([]byte("scope"))
			}
		}
		return nil
	})
}

func TestSearchRequest_ValidateInfo(t *testing.T) {
	Integration = 0
	SetUp()
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
			WriteOwnerInDB(tt.fields.Owner)
			if err := request.ValidateInfo(); err != nil {
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
		{name: "t1", fields: fields{Owner: "owner1", Token: "token1"}},
		{name: "t2", fields: fields{Owner: "owner2", Token: "token2"}},
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
			validators: tt.fields.validators,
		}
		if tt.name == "t1" {
			request.ValidateList()
			if request.Owner != "" && request.Token != "" {
				t.Errorf("%q. SearchRequest.ValidateList().Owner must be empty. Got: %v", tt.name, request.Owner)
			}
		} else if tt.name == "t2" {
			WriteOwnerInDB(tt.fields.Owner)
			request.ValidateList()
			if request.Owner != tt.fields.Owner {
				t.Errorf("%q. SearchRequest.ValidateList().Wait: %v, Got: %v", tt.name, tt.fields.Owner, request.Owner)
			}
		}
	}
}

func TestSearchRequest_ParseRequest(t *testing.T) {
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
		validators map[string]ValidateFunction
	}
	type args struct {
		httpRequest *http.Request
	}

	request1, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/raw/info?id=id1&owner=owner1&name=name1&version=version1&tags=tag1&token=token1&verified=false", nil)
	request2, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/apt/info?id=id2&owner=owner2&name=name2&version=version2&tags=tag2&token=token2&verified=false", nil)
	request3, _ := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/template/info?id=id3&owner=owner2&name=name2", nil)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "name1", Repo: "repo1", Version: "version1", Tags: "tag1",
			Token: "token1", Verified: "false", Operation: "list"}, args: args{httpRequest: request1}},
		{name: "t2", fields: fields{FileID: "id2", Owner: "owner2", Name: "name2", Repo: "repo2", Version: "version2", Tags: "tag2",
			Token: "token2", Verified: "false", Operation: "list"}, args: args{httpRequest: request2}},
		{name: "t3", fields: fields{FileID: "id3", Owner: "owner3", Name: "name3", Repo: "", Version: "", Tags: "",
			Token: "", Verified: "", Operation: ""}, args: args{httpRequest: request3}},
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
		if reflect.DeepEqual(SRequest, tt.fields) {
			t.Errorf("%q. Error. Wait %v, got %v", tt.name, tt.fields.FileID, SRequest.FileID)
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

func TestResult_GetResultByFileID(t *testing.T) {
	Integration = 0
	SetUp()
	defer TearDown()
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
	type args struct {
		fileID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "existedFile", Filename: "file1", Repo: "raw"}, args: args{fileID: "id1"}},
		{name: "t2", fields: fields{FileID: "id2", Owner: "owner2", Name: "NotExistedFile", Filename: "file2"}, args: args{fileID: "id2"}},
		{name: "t3", fields: fields{FileID: "id3", Owner: "owner3", Name: "TemplateFile", Filename: "file3-subutai-template", Repo: "template", Architecture: "arch"}, args: args{fileID: "id3"}},
		{name: "t4", fields: fields{FileID: "id4", Owner: "owner4", Name: "AptFile", Filename: "file4", Repo: "apt", Architecture: "arch"}, args: args{fileID: "id4"}},
		{name: "t5", fields: fields{FileID: "id5", Owner: "owner5", Name: "AptFile", Filename: "file5", Architecture: "arch"}, args: args{fileID: "id5"}},
		{name: "t6", fields: fields{FileID: "id6", Owner: "", Name: "AptFile", Filename: "file6", Repo: "apt", Architecture: "arch"}, args: args{fileID: "id6"}},
	}
	for _, tt := range tests {
		errored := false
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
		if tt.name == "t2" {
			result.GetResultByFileID(tt.args.fileID)
		}
		WriteOwnerInDB(result.Owner)
		WriteDB(result)
		result.GetResultByFileID(tt.args.fileID)
		if result.FileID != tt.fields.FileID &&
			result.Owner != tt.fields.Owner && result.Name != tt.fields.Name &&
			result.Filename != tt.fields.Filename && result.Repo != tt.fields.Repo &&
			result.Version != tt.fields.Version && result.Scope != tt.fields.Scope &&
			result.Md5 != tt.fields.Md5 && result.Sha256 != tt.fields.Sha256 &&
			result.Size != tt.fields.Size && result.Tags != tt.fields.Tags && result.Description != tt.fields.Description &&
			result.Architecture != tt.fields.Architecture && result.Parent != tt.fields.Parent &&
			result.ParentVersion != tt.fields.ParentVersion && result.ParentOwner != tt.fields.ParentOwner &&
			result.PrefSize != tt.fields.PrefSize {
			errored = true
			t.Errorf("%q. Result.GetResultByFileID() = %v, want %v", tt.name, result, tt.fields)
		}
		if errored {
			break
		}
	}
	TearDown()
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
		{name: "t3", fields: fields{Name: "file1", Owner: "subutai", Repo: "raw", Operation: "list"}},
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
