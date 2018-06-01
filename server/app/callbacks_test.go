package app

import (
	"net/http"
	"testing"
	"github.com/subutai-io/cdn/config"
	"os"
	"net/http/httptest"
	"github.com/subutai-io/cdn/libgorjun"
	"io/ioutil"
	"github.com/subutai-io/agent/log"
	"strings"
	"fmt"
	"github.com/subutai-io/cdn/db"
	"encoding/json"
)

func Clean() {
	os.Remove(config.DB.Path)
	files, _ := ioutil.ReadDir(config.Storage.Path)
	for _, file := range files {
		os.Remove(config.Storage.Path + file.Name())
	}
}

func SetUp() {
	log.Info("Setting up testing environment and configuration")
	Clean()
	InitFilters()
	config.DB.Path = "/tmp/data/db/my.db"
	config.Network.Port = "8080"
	config.Storage.Path = "/tmp/data/files/"
	config.Storage.Userquota = "2G"
	log.Info("Testing environment and configuration are set up")
	ListenAndServe()
	db.DB = db.InitDB()
}

func PreUploadUser(user gorjun.GorjunUser) (fileIDs [][]string) {
	publicFiles, _ := ioutil.ReadDir("/tmp/data/public/" + user.Username + "/")
	publicIDs := make([]string, 0)
	for _, file := range publicFiles {
		path := "/tmp/data/public/" + user.Username + "/" + file.Name()
		repo := ""
		if strings.Contains(file.Name(), "-subutai-template") {
			repo = "template"
			fileID, err := user.Upload(path, "template", "false")
			if err != nil {
				log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
			} else {
				publicIDs = append(publicIDs, fileID)
			}
		} else if strings.HasSuffix(file.Name(), ".deb") {
			repo = "apt"
			fileID, err := user.Upload(path, "apt", "false")
			if err != nil {
				log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
			} else {
				publicIDs = append(publicIDs, fileID)
			}
		} else {
			repo = "raw"
			fileID, err := user.Upload(path, "raw", "false")
			if err != nil {
				log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
			} else {
				publicIDs = append(publicIDs, fileID)
			}
		}
		log.Info(fmt.Sprintf("Uploading public file %s of user %s to repo %s", path, user.Username, repo))
	}
	fileIDs = append(fileIDs, publicIDs)
	log.Info(fmt.Sprintf("Public files of user %s are pre-uploaded to CDN", user.Username))
	privateFiles, _ := ioutil.ReadDir("/tmp/data/private/" + user.Username + "/")
	privateIDs := make([]string, 0)
	for _, file := range privateFiles {
		path := "/tmp/data/private/" + user.Username + "/" + file.Name()
		repo := ""
		if strings.Contains(file.Name(), "-subutai-template") {
			repo = "template"
			fileID, err := user.Upload(path, "template", "true")
			if err != nil {
				log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
			} else {
				privateIDs = append(privateIDs, fileID)
			}
		} else if strings.HasSuffix(file.Name(), ".deb") {
			repo = "apt"
			fileID, err := user.Upload(path, "apt", "true")
			if err != nil {
				log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
			} else {
				privateIDs = append(privateIDs, fileID)
			}
		} else {
			repo = "raw"
			fileID, err := user.Upload(path, "raw", "true")
			if err != nil {
				log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
			} else {
				privateIDs = append(privateIDs, fileID)
			}
		}
		log.Info(fmt.Sprintf("Uploading private file %s of user %s to repo %s", path, user.Username, repo))
	}
	fileIDs = append(fileIDs, privateIDs)
	log.Info(fmt.Sprintf("Private files of user %s are pre-uploaded to CDN", user.Username))
	log.Info(fmt.Sprintf("All uploaded files of user %s: %+v", user.Username, fileIDs))
	return
}

var (
	subutaiFiles      [][]string
	akenzhalievFiles  [][]string
	abaytulakovaFiles [][]string
)

func PreUpload() {
	log.Info("Pre-uploading files to CDN")
	subutai := gorjun.VerifiedGorjunUser()
	akenzhaliev := gorjun.FirstGorjunUser()
	abaytulakova := gorjun.SecondGorjunUser()
	subutaiFiles = PreUploadUser(subutai)
	akenzhalievFiles = PreUploadUser(akenzhaliev)
	abaytulakovaFiles = PreUploadUser(abaytulakova)
	log.Info("Pre-uploading files finished")
}

func TearDown() {
	Stop <- true
	<-Stop
	close(Stop)
	log.Info("Destroying testing environment")
	Clean()
	log.Info("Testing environment destroyed")
}

func TestFileSearch(t *testing.T) {
	SetUp()
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
		{name: "TestFileSearch-1",},
		{name: "TestFileSearch-2",},
		{name: "TestFileSearch-3",},
		// TODO: Add test cases.
	}
	for i := range tests {
		tt := &tests[i]
		if tt.name == "TestFileSearch-1" {
			tt.args.r, _ = http.NewRequest("GET", "/kurjun/rest/apt/list", nil)
		} else if tt.name == "TestFileSearch-2" {
			tt.args.r, _ = http.NewRequest("GET", "/kurjun/rest/raw/list", nil)
		} else if tt.name == "TestFileSearch-3" {
			tt.args.r, _ = http.NewRequest("GET", "/kurjun/rest/template/list", nil)
		}
	}
	for _, tt := range tests {
		errored := false
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(FileSearch)
			handler.ServeHTTP(recorder, tt.args.r)
			if tt.name == "TestFileSearch-1" {
				results := []*Result{new(Result)}
				log.Info("Expected abaytulakova's fileID: %s", abaytulakovaFiles[0][0])
				results[0].GetResultByFileID(abaytulakovaFiles[0][0])
				log.Info(fmt.Sprintf("Expected results[0]: %+v", results[0]))
				expected, err := json.Marshal(results)
				if err != nil {
					errored = true
					t.Errorf("%s didn't pass: %v", tt.name, err)
				} else if string(expected) != recorder.Body.String() {
					errored = true
					t.Errorf("%s didn't pass: unexpected result", tt.name)
				} else {
					log.Info("%s passed", tt.name)
				}
			} else if tt.name == "TestFileSearch-2" {
				results := []*Result{}
				log.Info(fmt.Sprintf("Expected results: %+v", results))
				expected, err := json.Marshal(results)
				if err != nil {
					errored = true
					t.Errorf("%s didn't pass: %v", tt.name, err)
				} else if string(expected) != recorder.Body.String() {
					errored = true
					t.Errorf("%s didn't pass: unexpected result", tt.name)
				} else {
					log.Info("%s passed", tt.name)
				}
			} else if tt.name == "TestFileSearch-3" {
				results := []*Result{new(Result), new(Result)}
				results[0].GetResultByFileID(subutaiFiles[0][0])
				results[1].GetResultByFileID(akenzhalievFiles[0][0])
				log.Info(fmt.Sprintf("Expected results[0]: %+v", results[0]))
				log.Info(fmt.Sprintf("Expected results[1]: %+v", results[1]))
				expected, err := json.Marshal(results)
				if err != nil {
					errored = true
					t.Errorf("%s didn't pass: %v", tt.name, err)
				} else if string(expected) != recorder.Body.String() {
					errored = true
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

func TestFileUpload(t *testing.T) {
	SetUp()
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "TestFileUpload-1",},
		{name: "TestFileUpload-2",},
		{name: "TestFileUpload-3",},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(FileUpload)
			handler.ServeHTTP(recorder, tt.args.r)
			if tt.name == "TestFileUpload-1" {
			} else if tt.name == "TestFileUpload-2" {
			} else if tt.name == "TestFileUpload-3" {
			}
		})
	}
	TearDown()
}
