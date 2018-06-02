package app

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"

	"github.com/subutai-io/cdn/db"
)

func WriteOwnerInDB(owner string) {
	db.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(db.Users); b != nil {
			b.CreateBucket([]byte(owner))
		}
		return nil
	})
}

func WriteFileInDB(fileID, fileName string) {
	db.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(db.MyBucket); b != nil {
			if c, _ := b.CreateBucket([]byte(fileID)); c != nil {
				c.Put([]byte("name"), []byte(fileName))
			}
		}
		return nil
	})
}

func TestSearchRequest_ValidateInfo(t *testing.T) {
	SetUp(false)
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
		{name: "t3", fields: fields{FileID: "id3", Owner: "ExistingOwner", Name: "name3"}},
		{name: "t4", fields: fields{FileID: "id4", Owner: "NotExistingOwner", Name: "name4"}},
		{name: "t5", fields: fields{FileID: "id5", Owner: "Owner", Name: "name5", Token: "token1"}},
		{name: "t6", fields: fields{FileID: "id6", Owner: "subutai", Name: "name6"}},
	}
	for _, tt := range tests {
		errorred := false
		request := &SearchRequest{
			FileID:   tt.fields.FileID,
			Owner:    tt.fields.Owner,
			Name:     tt.fields.Name,
			Token:    tt.fields.Token,
			Verified: tt.fields.Verified,
		}
		if tt.name == "t1" {
			if err := request.ValidateInfo(); err == nil {
				errorred = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		} else if tt.name == "t2" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name)
			if err := request.ValidateInfo(); err != nil {
				errorred = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		} else if tt.name == "t3" {
			WriteOwnerInDB(tt.fields.Owner)
			request.ValidateInfo()
			if request.Owner != tt.fields.Owner {
				errorred = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Owner is different. Wait: %v, got: %v", tt.name, tt.fields.Name, request.Owner)
			}
		} else if tt.name == "t4" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name)
			request.ValidateInfo()
			if request.Owner != "" {
				t.Errorf("%q. SearchRequest.ValidateInfo. Owner is not exist and must be empty. Got: %v", tt.name, request.Owner)
			}
		} else if tt.name == "t6" {
			WriteFileInDB(tt.fields.FileID, tt.fields.Name)
			WriteOwnerInDB(tt.fields.Owner)
			if err := request.ValidateInfo(); err != nil {
				errorred = true
				t.Errorf("%q. SearchRequest.ValidateInfo. Returned error: %v", tt.name, err)
			}
		}
		if errorred {
			break
		}
	}
	TearDown(false)
}

func TestSearchRequest_ValidateList(t *testing.T) {
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
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
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
		if err := request.ValidateList(); (err != nil) != tt.wantErr {
			t.Errorf("%q. SearchRequest.ValidateList() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSearchRequest_ParseRequest(t *testing.T) {
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

	request1, _ := http.NewRequest("POST", "/kurjun/rest/raw/info?id=id1&owner=owner1&name=name1&version=version1&tags=tag1&token=token1&verified=false", nil)
	request2, _ := http.NewRequest("POST", "/kurjun/rest/apt/info?id=id2&owner=owner2&name=name2&version=version2&tags=tag2&token=token2&verified=false", nil)
	request3, _ := http.NewRequest("POST", "/kurjun/rest/template/info?id=id3&owner=owner2&name=name2", nil)

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

	wantQuery1 := map[string]string{"FileID": "id1", "Owner": "owner1", "Name": "name1", "Repo": "repo1", "Version": "v1", "Tags": "tag1", "Verified": "true", "Operation": "info"}

	tests := []struct {
		name      string
		fields    fields
		wantQuery map[string]string
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "name1", Repo: "repo1", Version: "v1", Tags: "tag1", Verified: "true", Operation: "info"}, wantQuery: wantQuery1},
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
		if gotQuery := request.BuildQuery(); !reflect.DeepEqual(gotQuery, tt.wantQuery) {
			t.Errorf("%q. SearchRequest.BuildQuery() = %v, want %v", tt.name, gotQuery, tt.wantQuery)
		}
	}
}

func TestResult_ConvertToOld(t *testing.T) {
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

	//oldRes:=OldResult{FileID: "id1", Owner: "owner1", Name: "name1", Repo: "repo1", Version: "v1",Scope:"scope1", Tags: "tag1", Verified: "true", Operation: "info"}

	o := new(OldResult)
	tests := []struct {
		name   string
		fields fields
		want   *OldResult
	}{
		{name: "t1", fields: fields{FileID: "id1", Owner: "owner1", Name: "name1", Repo: "repo1", Version: "v1", Scope: "scope1", Tags: "tag1", Description: "description", Architecture: "amd64"}, want: o},
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
		info map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
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
		result.BuildResult(tt.args.info)
	}
}

func TestResult_GetResultByFileID(t *testing.T) {
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
		// TODO: Add test cases.
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
		result.GetResultByFileID(tt.args.fileID)
	}
}

func TestInitFilters(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for range tests {
		InitFilters()
	}
}

func TestFilterInfo(t *testing.T) {
	type args struct {
		query map[string]string
		files []*Result
	}
	tests := []struct {
		name       string
		args       args
		wantResult []*Result
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if gotResult := FilterInfo(tt.args.query, tt.args.files); !reflect.DeepEqual(gotResult, tt.wantResult) {
			t.Errorf("%q. FilterInfo() = %v, want %v", tt.name, gotResult, tt.wantResult)
		}
	}
}

func TestFilterList(t *testing.T) {
	type args struct {
		query map[string]string
		files []*Result
	}
	tests := []struct {
		name        string
		args        args
		wantResults []*Result
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if gotResults := FilterList(tt.args.query, tt.args.files); !reflect.DeepEqual(gotResults, tt.wantResults) {
			t.Errorf("%q. FilterList() = %v, want %v", tt.name, gotResults, tt.wantResults)
		}
	}
}

func TestSearchRequest_Retrieve(t *testing.T) {
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
		want   []*Result
	}{
		// TODO: Add test cases.
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
		if got := request.Retrieve(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. SearchRequest.Retrieve() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestGetFileInfo(t *testing.T) {
	type args struct {
		fileID string
	}
	tests := []struct {
		name     string
		args     args
		wantInfo map[string]string
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		gotInfo, err := GetFileInfo(tt.args.fileID)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. GetFileInfo() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
			t.Errorf("%q. GetFileInfo() = %v, want %v", tt.name, gotInfo, tt.wantInfo)
		}
	}
}

func TestMatchQuery(t *testing.T) {
	type args struct {
		file  map[string]string
		query map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := MatchQuery(tt.args.file, tt.args.query); got != tt.want {
			t.Errorf("%q. MatchQuery() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSearch(t *testing.T) {
	type args struct {
		query map[string]string
	}
	tests := []struct {
		name     string
		args     args
		wantList []*Result
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		gotList, err := Search(tt.args.query)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Search() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotList, tt.wantList) {
			t.Errorf("%q. Search() = %v, want %v", tt.name, gotList, tt.wantList)
		}
	}
}
