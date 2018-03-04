package gorjun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// GorjunServer is a representation for Gorjun bucket
type GorjunServer struct {
	Username     string // Username of gorjun user
	Email        string // Email used to identify user in GPG
	Hostname     string // Hostname of the Gorjun server
	GPGDirectory string // GPGDirectory points to a gnupg directory in the file system
	Token        string // Active token
	TokenCode    string // Clean token code
	Passphrase   string // Passphrase used to decrypt private key
}

func NewGorjunServer() GorjunServer {
	return GorjunServer{"tester", "tester@gmail.com", "127.0.0.1:8080", os.Getenv("HOME") + "/.gnupg",
		"", "", "pantano"}
}

// GorjunFile is a file located on Gorjun bucket server
type GorjunFile struct {
	ID           string            `json:"id"`
	Hash         hashsums          `json:"hash"`
	Size         int               `json:"size"`
	Date         time.Time         `json:"upload-date-formatted"`
	Timestamp    string            `json:"upload-date-timestamp,omitempty"`
	Name         string            `json:"name,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Owner        []string          `json:"owner,omitempty"`
	Parent       string            `json:"parent,omitempty"`
	Version      string            `json:"version,omitempty"`
	Filename     string            `json:"filename,omitempty"`
	Prefsize     string            `json:"prefsize,omitempty"`
	Signature    map[string]string `json:"signature,omitempty"`
	Description  string            `json:"description,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
}
type Keys struct {
}
type hashsums struct {
	Md5    string `json:"md5,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

// ListUserFiles returns a list of files that belongs to user
func (g *GorjunServer) ListUserFiles() ([]GorjunFile, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/raw/info?owner=%s", g.Hostname, g.Username))
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve file list from %s: %v", g.Hostname, err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	var rf []GorjunFile
	err = json.Unmarshal(data, &rf)
	if err != nil {
		log.Printf("error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", data)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal contents from %s: %v", g.Hostname, err)
	}
	return rf, nil
}

// GetFileByName will return information about a file with specified name
func (g *GorjunServer) GetFileByName(filename string, artifactType string) ([]GorjunFile, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/%s/info?name=%s&owner=%s", g.Hostname, artifactType, filename, g.Username))
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve file information from %s: %v", g.Hostname, err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	var f []GorjunFile
	err = json.Unmarshal(data, &f)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal contents from %s: %v", g.Hostname, err)
	}
	return f, nil
}

// UploadFile will upload file and return it's ID after successful upload
func (g *GorjunServer) Upload(filename string, artifactType string) (string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("%s not found", filename)
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("Failed to open file: %v", err)
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("Failed to create file form: %v", err)
	}
	if _, err = io.Copy(fw, f); err != nil {
		return "", fmt.Errorf("Failed to copy file contents: %v", err)
	}
	if fw, err = w.CreateFormField("token"); err != nil {
		return "", fmt.Errorf("Failed to create token form field: %v", err)
	}
	if _, err = fw.Write([]byte(g.Token)); err != nil {
		return "", fmt.Errorf("Failed to write token: %v", err)
	}

	w.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/kurjun/rest/"+artifactType+"/upload", g.Hostname), &b)
	if err != nil {
		return "", fmt.Errorf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("token", g.Token)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to execute HTTP request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Upload failed. Server returned %s error", res.Status)
	}
	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}
	return string(response), nil
}

// RemoveFile will delete file on gorjun with specified name. If multiple files with the same
// name exists belong to the same user only the last one (most recent) will be removed
func (g *GorjunServer) RemoveFile(filename string, artifactType string) error {
	file, err := g.GetFileByName(filename, artifactType)
	if err != nil {
		return fmt.Errorf("Failed to get file: %v", err)
	}
	fmt.Printf("\nId of artifact with type %s is %s going to deleted", artifactType, file[0].ID)
	return g.RemoveFileByID(file[0].ID, "raw")
}

// RemoveFileByID will remove file with specified ID
func (g *GorjunServer) RemoveFileByID(ID string, artifactType string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s/kurjun/rest/%s/delete?id=%s&token=%s",
		g.Hostname, artifactType, ID, g.Token), nil)

	if err != nil {
		return fmt.Errorf("Failed to remove file [%s]: %s", ID, err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to remove file: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Can't remove file - HTTP request returned %s code", res.Status)
	}
	fmt.Printf("\nId of artifact with type %s is %s deleted\n", artifactType, ID)
	fmt.Printf("\n%s\n", req.URL)
	return nil
}

// DownloadFile will download file with specified name into the specified output directory
func (g *GorjunServer) DownloadFile(filename, outputDirectory string) error {
	return nil
}

// DownloadFileByID will download file with specified ID into the specified output directory
func (g *GorjunServer) DownloadFileByID(ID, outputDirectory string) error {
	return nil
}
