package app

import (
	"testing"
	"net/http"
	"strconv"
	"net/http/httptest"
	"encoding/json"
	"github.com/subutai-io/agent/log"
	"fmt"
)

func TestFileSearch(t *testing.T) {
	Integration = 1
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	PreUpload()
	defer TearDown()
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "TestFileSearch-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 2; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	tests[0].args.r, _ = http.NewRequest("GET", "http://127.0.0.1:8080/kurjun/rest/apt/list", nil)
	tests[1].args.r, _ = http.NewRequest("GET", "http://127.0.0.1:8080/kurjun/rest/raw/list", nil)
	tests[2].args.r, _ = http.NewRequest("GET", "http://127.0.0.1:8080/kurjun/rest/template/list", nil)
	for i, tt := range tests {
		errored := false
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(FileSearch)
			handler.ServeHTTP(recorder, tt.args.r)
			var err error
			var expected []byte
			results := make([]*Result, 0)
			oldResults := make([]*OldResult, 0)
			if i == 0 {
				results = append(results, new(Result))
				results[0].GetResultByFileID(UserFiles[PublicScope][Lorem.Username][2])
			} else if i == 1 || i == 2 {
				results = append(results, new(Result))
				results = append(results, new(Result))
				results = append(results, new(Result))
				if i == 1 {
					results[0].GetResultByFileID(UserFiles[PublicScope][Lorem.Username][i - 1])
					results[1].GetResultByFileID(UserFiles[PublicScope][Subutai.Username][i - 1])
					results[2].GetResultByFileID(UserFiles[PublicScope][Ipsum.Username][i - 1])
				} else {
					results[0].GetResultByFileID(UserFiles[PublicScope][Lorem.Username][i - 1])
					results[1].GetResultByFileID(UserFiles[PublicScope][Subutai.Username][i - 1])
					results[2].GetResultByFileID(UserFiles[PublicScope][Ipsum.Username][i - 1])
				}
			}
			for i := 0; i < len(results); i++ {
				oldResults = append(oldResults, results[i].ConvertToOld())
			}
			expected, err = json.Marshal(oldResults)
			if err != nil {
				errored = true
				t.Errorf("%s didn't pass: %v", tt.name, err)
			} else {
				list := make([]*OldResult, len(oldResults))
				expectedList := make([]*OldResult, len(oldResults))
				json.Unmarshal(recorder.Body.Bytes(), list)
				json.Unmarshal(expected, expectedList)
				if !SlicesEqual(list, expectedList) {
					errored = true
					log.Warn(fmt.Sprintf("%s didn't pass: unexpected result:", tt.name))
					log.Warn(fmt.Sprintf("%+v", list))
					log.Warn(fmt.Sprintf("%+v", expectedList))
					t.Errorf("%s didn't pass: unexpected result", tt.name)
				} else {
					log.Info("%s passed", tt.name)
				}
			}
		})
		if errored {
			break
		}
	}
	log.Info("TestFileSearch ended")
}
