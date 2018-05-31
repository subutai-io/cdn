package app

import (
	"net/http"
	"testing"
	"github.com/subutai-io/cdn/config"
	"os"
	"net/http/httptest"
	"github.com/subutai-io/cdn/libgorjun"
	"io/ioutil"
	"strings"
	"github.com/subutai-io/agent/log"
)

func Clean() {
	os.Remove(config.DB.Path)
	os.RemoveAll(config.Storage.Path)
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
}

func PreUpload() {
	log.Info("Pre-uploading files to CDN")
	subutai := gorjun.VerifiedGorjunUser()
	publicFiles, _ := ioutil.ReadDir("/tmp/data/public/")
	for _, file := range publicFiles {
		path := config.Storage.Path + file.Name()
		if strings.Contains(file.Name(), "-subutai-template") {
			subutai.Upload(path, "template", "false")
		} else if strings.HasSuffix(file.Name(), ".deb") {
			subutai.Upload(path, "apt", "false")
		} else {
			subutai.Upload(path, "raw", "false")
		}
	}
	log.Info("Public files are pre-uploaded files to CDN")
	privateFiles, _ := ioutil.ReadDir("/tmp/data/private/")
	for _, file := range privateFiles {
		path := config.Storage.Path + file.Name()
		if strings.Contains(file.Name(), "-subutai-template") {
			subutai.Upload(path, "template", "true")
		} else if strings.HasSuffix(file.Name(), ".deb") {
			subutai.Upload(path, "apt", "true")
		} else {
			subutai.Upload(path, "raw", "true")
		}
	}
	log.Info("Private files are pre-uploaded files to CDN")
}

func TearDown() {
	log.Info("Destroying testing environment")
	Clean()
	log.Info("Testing environment destroyed")
}

func TestFileSearch(t *testing.T) {
	SetUp()
	PreUpload()
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
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(FileSearch)
			handler.ServeHTTP(recorder, tt.args.r)
			if tt.name == "TestFileSearch-1" {
			} else if tt.name == "TestFileSearch-2" {
			} else if tt.name == "TestFileSearch-3" {
			}
		})
	}
	log.Info("TestFileSearch ended")
	TearDown()
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
