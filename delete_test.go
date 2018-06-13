package app

import (
	"net/http"
	"testing"
	"strconv"
	"fmt"
)

func TestDeleteRequest_ValidateRequest(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	PreUpload()
	defer TearDown()
	type fields struct {
		FileID string
		Owner  string
		Token  string
		Repo   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "TestDeleteRequest_ValidateRequest-"},
		// TODO: Add test cases.
	}
	for i := 0; i <= 2; i++ {
		test := tests[0]
		test.fields = fields{
			FileID: UserFiles[PublicScope][Lorem.Username][i],
			Token:  Lorem.Token,
		}
		test.wantErr = false
		if i > 0 {
			test.name += strconv.Itoa(i)
			tests = append(tests, test)
		} else {
			tests[i] = test
		}
	}
	fileIDs := []string{UserFiles[PublicScope][Lorem.Username][3], ""}
	tokens := []string{"", Lorem.Token}
	for i := 3; i <= 4; i++ {
		test := tests[0]
		test.fields = fields{
			FileID: fileIDs[i % 3],
			Token:  tokens[i % 3],
		}
		test.wantErr = true
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &DeleteRequest{
				FileID: tt.fields.FileID,
				Owner:  tt.fields.Owner,
				Token:  tt.fields.Token,
				Repo:   tt.fields.Repo,
			}
			if err := request.ValidateRequest(); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRequest.ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteRequest_ParseRequest(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	PreUpload()
	defer TearDown()
	type fields struct {
		FileID string
		Owner  string
		Token  string
		Repo   string
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "TestDeleteRequest_ParseRequest-"},
		// TODO: Add test cases.
	}
	for i := 0; i <= 2; i++ {
		test := tests[0]
		request, _ := http.NewRequest("", fmt.Sprintf("http://127.0.0.1:8080/kurjun/rest/%s/delete?id=%s&token=%s", Repos[PublicScope][Lorem.Username][i], UserFiles[PublicScope][Lorem.Username][i], Lorem.Token), nil)
		test.args = args{
			r: request,
		}
		test.wantErr = false
		if i > 0 {
			test.name += strconv.Itoa(i)
			tests = append(tests, test)
		} else {
			tests[i] = test
		}
	}
	fileIDs := []string{UserFiles[PublicScope][Lorem.Username][3], ""}
	tokens := []string{"", Lorem.Token}
	for i := 3; i <= 4; i++ {
		test := tests[0]
		request, _ := http.NewRequest("", fmt.Sprintf("http://127.0.0.1:8080/kurjun/rest/%s/delete?id=%s&token=%s", Repos[PublicScope][Lorem.Username][i], fileIDs[i % 3], tokens[i % 3]), nil)
		test.args = args{
			r: request,
		}
		test.wantErr = true
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &DeleteRequest{
				FileID: tt.fields.FileID,
				Owner:  tt.fields.Owner,
				Token:  tt.fields.Token,
				Repo:   tt.fields.Repo,
			}
			if err := request.ParseRequest(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRequest.ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteRequest_Delete(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	PreUpload()
	defer TearDown()
	type fields struct {
		FileID string
		Owner  string
		Token  string
		Repo   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "TestDeleteRequest_Delete-"},
		// TODO: Add test cases.
	}
	for i := 0; i <= 2; i++ {
		test := tests[0]
		test.fields = fields{
			FileID: UserFiles[PublicScope][Lorem.Username][i],
			Owner:  Lorem.Username,
			Token:  Lorem.Token,
			Repo:   Repos[PublicScope][Lorem.Username][i],
		}
		test.wantErr = false
		if i > 0 {
			test.name += strconv.Itoa(i)
			tests = append(tests, test)
		} else {
			tests[i] = test
		}
	}
	{
		test := tests[0]
		test.fields = fields{
			FileID: UserFiles[PublicScope][Lorem.Username][3],
			Owner:  Lorem.Username,
			Token:  Lorem.Token,
			Repo:   "LoremIpsumDolorSitAmet",
		}
		test.wantErr = true
		test.name += "3"
		tests = append(tests, test)
	}
	for i := 4; i <= 5; i++ {
		test := tests[0]
		test.fields = fields{
			FileID: UserFiles[PublicScope][Ipsum.Username][i % 2],
			Owner:  Ipsum.Username,
			Token:  Ipsum.Token,
			Repo:   Repos[PublicScope][Ipsum.Username][i % 2],
		}
		test.wantErr = false
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	for i := 6; i <= 7; i++ {
		test := tests[0]
		test.fields = fields{
			FileID: UserFiles[PublicScope][Subutai.Username][i % 2],
			Owner:  Subutai.Username,
			Token:  Subutai.Token,
			Repo:   Repos[PublicScope][Subutai.Username][i % 2],
		}
		test.wantErr = false
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &DeleteRequest{
				FileID: tt.fields.FileID,
				Owner:  tt.fields.Owner,
				Token:  tt.fields.Token,
				Repo:   tt.fields.Repo,
			}
			if err := request.Delete(); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRequest.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteFS(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	PreUpload()
	defer TearDown()
	type args struct {
		result *Result
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteFS(tt.args.result)
		})
	}
}
