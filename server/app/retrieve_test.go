package app

import (
	"net/http"
	"reflect"
	"testing"
)

func TestSearchRequest_InitValidators(t *testing.T) {
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
		request.InitValidators()
	}
}

func TestSearchRequest_ValidateRequest(t *testing.T) {
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
		if err := request.ValidateRequest(); (err != nil) != tt.wantErr {
			t.Errorf("%q. SearchRequest.ValidateRequest() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSearchRequest_ValidateInfo(t *testing.T) {
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
		if err := request.ValidateInfo(); (err != nil) != tt.wantErr {
			t.Errorf("%q. SearchRequest.ValidateInfo() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
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
	tests := []struct {
		name    string
		fields  fields
		args    args
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
		if err := request.ParseRequest(tt.args.httpRequest); (err != nil) != tt.wantErr {
			t.Errorf("%q. SearchRequest.ParseRequest() error = %v, wantErr %v", tt.name, err, tt.wantErr)
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
	tests := []struct {
		name      string
		fields    fields
		wantQuery map[string]string
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
	tests := []struct {
		name   string
		fields fields
		want   *OldResult
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
