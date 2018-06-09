package app

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/apt"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
)

type Hashes struct {
	Md5    string `json:"md5,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

type OldResult struct {
	FileID        string   `json:"id,omitempty"`
	Owner         []string `json:"owner,omitempty"`
	Name          string   `json:"name,omitempty"`
	Filename      string   `json:"filename,omitempty"`
	Version       string   `json:"version,omitempty"`
	Hash          Hashes   `json:"hash,omitempty"`
	Size          int64    `json:"size,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Date          string   `json:"upload-date-formatted,omitempty"`
	Timestamp     string   `json:"upload-date-timestamp,omitempty"`
	Description   string   `json:"description,omitempty"`
	Architecture  string   `json:"architecture,omitempty"`
	Parent        string   `json:"parent,omitempty"`
	ParentVersion string   `json:"parent-version,omitempty"`
	ParentOwner   string   `json:"parent-owner,omitempty"`
	PrefSize      string   `json:"prefsize,omitempty"`
}

// FileSearch handles the info and list HTTP requests
func FileSearch(w http.ResponseWriter, r *http.Request) {
	log.Info("Received FileSearch request")
	log.Info(r)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info/list request"))
		return
	}
	request := new(SearchRequest)
	request.InitValidators()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect info/list request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	files := request.Retrieve()
	log.Info("Successfully retrieved files")
	oldFiles := make([]*OldResult, 0)
	for _, file := range files {
		oldFiles = append(oldFiles, file.ConvertToOld())
	}
	log.Info("Retrieve: ", oldFiles)
	log.Info("Successfully converted files to old format")
	result, _ := json.Marshal(oldFiles)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
	log.Info("Successfully handled FileSearch request")
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	log.Info("Received upload request")
	log.Info(r)
	if r.Method != "POST" {
		log.Warn("Incorrect method for upload request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect method for upload request")))
		return
	}
	request := new(UploadRequest)
	request.InitUploaders()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		log.Warn("Couldn't parse upload request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect upload request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	err = request.Upload()
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't upload the file: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while uploading the file: %v", err)))
		return
	}
	log.Info("Successfully uploaded a file: ", request.fileID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(request.fileID))
}

func FileDownload(w http.ResponseWriter, r *http.Request) {
	log.Info("Received download request")
	log.Info(r)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for download request"))
		return
	}
	request := new(DownloadRequest)
	request.InitDownloaders()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		log.Warn("Couldn't parse download request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect download request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	result, token, err := request.ExecRequest()
	if result.Repo == "raw" || result.Repo == "template" {
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		}
		path := config.ConfigurationStorage.Path + result.FileID
		if md5, _ := db.Hash(result.FileID); len(md5) != 0 {
			path = config.ConfigurationStorage.Path + md5
		}
		file, err := os.Open(path)
		defer file.Close()
		if log.Check(log.WarnLevel, "Opening file "+config.ConfigurationStorage.Path+result.FileID, err) || len(result.FileID) == 0 {
			if len(config.ConfigurationCDN.Node) > 0 {
				client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
				resp, err := client.Get(config.ConfigurationCDN.Node + r.URL.RequestURI())
				if !log.Check(log.WarnLevel, "Getting file from CDN", err) {
					w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
					w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
					w.Header().Set("Last-Modified", resp.Header.Get("Last-Modified"))
					w.Header().Set("Content-Disposition", resp.Header.Get("Content-Disposition"))
					io.Copy(w, resp.Body)
					resp.Body.Close()
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "File not found")
			return
		}
		fileInfo, _ := file.Stat()
		if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && fileInfo.ModTime().Unix() <= t.Unix() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(fileInfo.Size()))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))
		if name := db.NameByHash(result.FileID); len(name) == 0 && len(config.ConfigurationCDN.Node) > 0 {
			httpclient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			resp, err := httpclient.Get(config.ConfigurationCDN.Node + "/kurjun/rest/template/info?id=" + result.FileID + "&token=" + token)
			if !log.Check(log.WarnLevel, "Getting info from CDN", err) {
				info := new(Result)
				rsp, err := ioutil.ReadAll(resp.Body)
				if log.Check(log.WarnLevel, "Reading from CDN response", err) {
					w.WriteHeader(http.StatusNotFound)
					io.WriteString(w, "File not found")
					return
				}
				if !log.Check(log.WarnLevel, "Decrypting request", json.Unmarshal([]byte(rsp), &info)) {
					w.Header().Set("Content-Disposition", "attachment; filename=\""+info.Filename+"\"")
				}
				resp.Body.Close()
			}
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename=\""+db.NameByHash(result.FileID)+"\"")
		}
		io.Copy(w, file)
	} else if result.Repo == "apt" {
		file := result.Filename
		if len(file) == 0 {
			file = strings.TrimPrefix(r.RequestURI, "/kurjun/rest/apt/")
		}
		size := GetSize(config.ConfigurationStorage.Path + "Packages")
		if file == "Packages" && size == 0 {
			apt.GenerateReleaseFile()
		}
		if f, err := os.Open(config.ConfigurationStorage.Path + file); err == nil && file != "" {
			defer f.Close()
			stats, _ := f.Stat()
			w.Header().Set("Content-Length", strconv.Itoa(int(stats.Size())))
			io.Copy(w, f)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func FileDelete(w http.ResponseWriter, r *http.Request) {
	request := new(DeleteRequest)
	err := request.ParseRequest(r)
	if err != nil {
		log.Warn("Couldn't parse delete request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect delete request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	err = request.Delete()
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't delete the file: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while deleting the file: %v", err)))
		return
	}
	log.Info("Successfully deleted a file: ", request.FileID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(request.FileID))
}
