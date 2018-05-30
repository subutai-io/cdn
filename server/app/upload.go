package app

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"strings"
	"strconv"
	"net/http"
	"archive/tar"
	"compress/gzip"

	"github.com/satori/go.uuid"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/config"
	"regexp"
	"reflect"
	"github.com/subutai-io/cdn/download"
)

type UploadRequest struct {
	File     io.Reader
	Filename string
	Repo     string
	Owner    string
	Token    string
	Private  string
	Md5      string
	Sha256   string
	Size     int

	uploaders map[string]UploadFunction
}

func (request *UploadRequest) ParseRequest(r *http.Request) error {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		return err
	}
	request.File = io.Reader(file) // multipart.sectionReadCloser
	limit := int64(db.QuotaLeft(request.Owner))
	if limit != -1 {
		request.File = io.LimitReader(file, limit)
	}
	request.Filename = header.Filename
	request.Repo = strings.Split(r.RequestURI, "/")[3]
	request.Token = r.Header.Get("token")
	if len(request.Token) == 0 {
		return fmt.Errorf("token for upload wasn't provided")
	}
	request.Owner = db.TokenOwner(request.Token)
	if request.Owner == "" {
		return fmt.Errorf("incorrect token provided")
	}
	if len(r.MultipartForm.Value["private"]) > 0 {
		request.Private = r.MultipartForm.Value["private"][0]
	}
	return nil
}

type UploadFunction func(string, string, int64) error

func (request *UploadRequest) InitUploaders() {
	request.uploaders             = make(map[string]UploadFunction)
	request.uploaders["apt"]      = request.UploadApt
	request.uploaders["raw"]      = request.UploadRaw
	request.uploaders["template"] = request.UploadTemplate
}

func (request *UploadRequest) ExecRequest(md5Sum, sha256Sum string, size int64) error {
	uploader := request.uploaders[request.Repo]
	return uploader(md5Sum, sha256Sum, size)
}

func (request *UploadRequest) UploadApt(md5Sum, sha256Sum string, size int64) error {

	return nil
}

func (request *UploadRequest) UploadRaw(md5Sum, sha256Sum string, size int64) error {

	return nil
}

func LoadConfiguration(md5Sum string) (configuration string, err error) {
	var configurationFile bytes.Buffer
	f, err := os.Open(config.Storage.Path + md5Sum)
	if err != nil {
		return
	}
	defer f.Close()
	gzFile, err := gzip.NewReader(f)
	if err != nil {
		return
	}
	tarFile := tar.NewReader(gzFile)
	for file, fileErr := tarFile.Next(); fileErr != io.EOF; file, err = tarFile.Next() {
		if file.Name == "config" {
			if _, err = io.Copy(&configurationFile, tarFile); err != nil {
				return
			}
			break
		}
	}
	configuration = configurationFile.String()
	return
}

func FormatConfiguration(hash string, configuration string) (template *Result) {
	my_uuid, _ := uuid.NewV4()
	template = new(Result)
	template.FileID = my_uuid.String()
	template.Md5 = hash
	for _, line := range strings.Split(configuration, "\n") {
		if blocks := strings.Split(line, "="); len(blocks) > 1 {
			blocks[0] = strings.ToLower(strings.TrimSpace(blocks[0]))
			blocks[1] = strings.ToLower(strings.TrimSpace(blocks[1]))
			switch blocks[0] {
			case "lxc.arch":
				template.Architecture = blocks[1]
			case "lxc.utsname":
				template.Name = blocks[1]
			case "subutai.parent":
				template.Parent = blocks[1]
			case "subutai.parent.owner":
				template.ParentOwner = blocks[1]
			case "subutai.parent.version":
				template.ParentVersion = blocks[1]
			case "subutai.template.version":
				template.Version = blocks[1]
			case "subutai.template.size":
				template.PrefSize = blocks[1]
			case "subutai.template.owner":
				template.Owner = blocks[1]
			case "subutai.template.description":
				template.Description = blocks[1]
			case "subutai.tags":
				template.Tags = blocks[1]
			}
		}
	}
	return
}

func (request *UploadRequest) TemplateCheckValid(template *Result) error {
	err := request.TemplateCheckFieldsPresent(template)
	if err != nil {
		return err
	}
	err = request.TemplateCheckOwner(template)
	if err != nil {
		return err
	}
	err = request.TemplateCheckDependencies(template)
	if err != nil {
		return err
	}
	err = request.TemplateCheckFormat(template)
	if err != nil {
		return err
	}
	return nil
}

func (request *UploadRequest) TemplateCheckFieldsPresent(template *Result) error {
	s := reflect.ValueOf(template).Elem()
	typeOfT := s.Type()
	requiredFields := []string{"Parent", "ParentOwner", "ParentVersion", "Version", "Name", "Owner"}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fieldName := typeOfT.Field(i).Name
		fieldValue := f.Interface()
		if (download.In(fieldName, requiredFields) && fieldValue == "") ||
			(fieldName == "Owner" && len(template.Owner) == 0) {
			return fmt.Errorf("%s field required", fieldName)
		}
	}
	return nil
}

func (request *UploadRequest) TemplateCheckOwner(template *Result) error {
	if template.Owner != request.Owner {
		return fmt.Errorf("owner in config file is different")
	}
	return nil
}

func (request *UploadRequest) TemplateCheckDependencies(template *Result) error {
	parentExists := false
	list := db.SearchName(template.Parent)
	for _, id := range list {
		info, err := GetFileInfo(id)
		if err != nil {
			continue
		}
		item := new(Result)
		item.BuildResult(info)
		if len(item.Owner) == 0 {
			continue
		}
		if item.Name == template.Parent &&
			item.Owner == template.ParentOwner &&
			item.Version == template.ParentVersion {
			parentExists = true
			break
		}
	}
	if parentExists || template.Name == template.Parent {
		return nil
	}
	return fmt.Errorf("dependencies are not correct")
}

func (request *UploadRequest) TemplateCheckFormat(template *Result) error {
	name, _    := regexp.MatchString("^[a-zA-Z0-9._-]+$", template.Name)
	version, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", template.Version)
	if name && version {
		return nil
	}
	return fmt.Errorf("name or version format is wrong")
}

func (request *UploadRequest) UploadTemplate(md5Sum, sha256Sum string, size int64) error {
	configuration, err := LoadConfiguration(md5Sum)
	if err != nil {
		return err
	}
	result := FormatConfiguration(md5Sum, configuration)
	err = request.TemplateCheckValid(result)
	if err != nil {

	}
	return nil
}

func (request *UploadRequest) Upload() error {
	filePath := config.Storage.Path + request.Filename
	file, err := os.Create(filePath)
	if err != nil {
		log.Warn("Couldn't create file for writing - %s", filePath)
		return err
	}
	defer file.Close()
	limit := int64(db.QuotaLeft(request.Owner))
	size, err := io.Copy(file, request.File)
	if limit != -1 && (size == limit || err != nil) {
		log.Warn("User " + request.Owner + " exceeded storage quota, removing file")
		os.Remove(filePath)
		return fmt.Errorf("failed to write file or storage quota exceeded")
	} else {
		db.QuotaUsageSet(request.Owner, int(size))
		log.Info("User " + request.Owner + ", quota usage +" + strconv.Itoa(int(size)))
	}
	md5Sum    := Hash(filePath, "md5")
	sha256Sum := Hash(filePath, "sha256")
	if len(md5Sum) == 0 || len(sha256Sum) == 0 {
		log.Warn("Failed to calculate hash for " + request.Filename)
		return fmt.Errorf("failed to calculate hash")
	}
	if request.Repo != "apt" {
		md5Path := config.Storage.Path + md5Sum
		os.Rename(filePath, md5Path)
		log.Debug(fmt.Sprintf("repo is not apt: renamed %s to %s", filePath, md5Path))
	}
	return request.ExecRequest(md5Sum, sha256Sum, size)
}
