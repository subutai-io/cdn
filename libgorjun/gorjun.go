package gorjun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"github.com/subutai-io/agent/log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"strings"
)

// GorjunUser is a representation for Gorjun bucket
type GorjunUser struct {
	Username     string // Username of gorjun user
	Email        string // Email used to identify user in GPG
	Hostname     string // Hostname of the Gorjun server
	GPGDirectory string // GPGDirectory points to a gnupg directory in the file system
	Token        string // Active token
	TokenCode    string // Clean token code
	Passphrase   string // Passphrase used to decrypt private key
}

func FirstGorjunUser() GorjunUser {
	return GorjunUser{"akenzhaliev", "akenzhaliev@optimal-dynamics.com", "127.0.0.1:8080", os.Getenv("HOME") + "/.gnupg",
	"", "", ""}
}

func SecondGorjunUser() GorjunUser {
	return GorjunUser{"abaytulakova", "abaytulakova@optimal-dynamics.com", "127.0.0.1:8080", os.Getenv("HOME") + "/.gnupg",
	"", "", ""}
}

func VerifiedGorjunUser() GorjunUser {
	return GorjunUser{"subutai", "subutai@subutai.io", "127.0.0.1:8080", os.Getenv("HOME") + "/.gnupg",
	"", "", ""}
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
func (g *GorjunUser) ListUserFiles() ([]GorjunFile, error) {
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
		log.Debug("error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Debug("syntax error at byte offset %d", e.Offset)
		}
		log.Debug("sakura response: %q", data)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal contents from %s: %v", g.Hostname, err)
	}
	return rf, nil
}

// GetFileByName will return information about a file with specified name
func (g *GorjunUser) GetFileByName(filename string, artifactType string) ([]GorjunFile, error) {
	if artifactType == "template" {
		filename = strings.Split(filename, "-subutai-template")[0]
	}
	fmt.Println(fmt.Sprintf("http://%s/kurjun/rest/%s/info?name=%s&token=%s", g.Hostname, artifactType, filename, g.Token))
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/%s/info?name=%s&token=%s", g.Hostname, artifactType, filename, g.Token))
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
func (g *GorjunUser) Upload(filename string, artifactType string, private string) (string, error) {
	log.Info("Sending Upload request")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Warn(fmt.Sprintf("File %s not found", filename))
		return "", fmt.Errorf("%s not found", filename)
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	f, err := os.Open(filename)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to open file: %v", err))
		return "", fmt.Errorf("Failed to open file: %v", err)
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to create file form: %v", err))
		return "", fmt.Errorf("Failed to create file form: %v", err)
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Warn(fmt.Sprintf("Failed to copy file contents: %v", err))
		return "", fmt.Errorf("Failed to copy file contents: %v", err)
	}
	if fw, err = w.CreateFormField("token"); err != nil {
		log.Warn(fmt.Sprintf("Failed to create token form field: %v", err))
		return "", fmt.Errorf("Failed to create token form field: %v", err)
	}
	if _, err = fw.Write([]byte(g.Token)); err != nil {
		log.Warn(fmt.Sprintf("Failed to write token: %v", err))
		return "", fmt.Errorf("Failed to write token: %v", err)
	}
	if fw, err = w.CreateFormField("private"); err != nil {
		log.Warn(fmt.Sprintf("Failed to create private form field: %v", err))
		return "", fmt.Errorf("Failed to create private form field: %v", err)
	}
	if _, err = fw.Write([]byte(private)); err != nil {
		log.Warn(fmt.Sprintf("Failed to write private: %v", err))
		return "", fmt.Errorf("Failed to write private: %v", err)
	}
	w.Close()
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/kurjun/rest/%s/upload", g.Hostname, artifactType), &b)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to create HTTP request: %v", err))
		return "", fmt.Errorf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("token", g.Token)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to execute HTTP request: %v", err))
		return "", fmt.Errorf("Failed to execute HTTP request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		log.Warn(fmt.Sprintf("Upload failed. Server returned %s error", res.Status))
		return "", fmt.Errorf("Upload failed. Server returned %s error", res.Status)
	}
	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to read response body: %v", err))
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}
	return string(response), nil
}

// RemoveFile will delete file from gorjun with specified name. If multiple files with the same
// name exists belong to the same user only the last one (most recent) will be removed
func (g *GorjunUser) RemoveFile(filename string, artifactType string) error {
	file, err := g.GetFileByName(filename, artifactType)
	if err != nil {
		return fmt.Errorf("Failed to get file: %v", err)
	}
	fmt.Printf("\nID %+v of artifact with type %+v is going to deleted", file[0].ID, artifactType)
	return g.RemoveFileByID(file[0].ID, "raw")
}

// RemoveFileByID will remove file with specified ID
func (g *GorjunUser) RemoveFileByID(id string, artifactType string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s/kurjun/rest/%s/delete?id=%s&token=%s", g.Hostname, artifactType, id, g.Token), nil)
	if err != nil {
		return fmt.Errorf("Failed to remove file [%+v] (1): %+v", id, err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to remove file [%+v] (2): %+v", id, err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Can't remove file: HTTP request returned %+v code", res.Status)
	}
	fmt.Printf("\nArtifact with ID %+v is deleted from %+v repo\n", id, artifactType)
	fmt.Printf("\n%+v\n", req.URL)
	fmt.Printf("\nResponse from gorjun = ")
	io.Copy(os.Stdout, res.Body)
	fmt.Println()
	return nil
}

// DownloadFile will download file with specified name into the specified output directory
func (g *GorjunUser) DownloadFile(filename, outputDirectory string) error {
	return nil
}

// DownloadFileByID will download file with specified ID into the specified output directory
func (g *GorjunUser) DownloadFileByID(ID, outputDirectory string) error {
	return nil
}