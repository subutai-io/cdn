package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	ar "github.com/mkrautz/goar"
	uuid "github.com/satori/go.uuid"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/cdn/download"
)

type UploadRequest struct {
	File     io.Reader
	Filename string
	Repo     string
	Owner    string
	Token    string
	Private  string
	Tags     string
	Version  string

	fileID string
	md5    string
	sha256 string
	size   int64

	uploaders map[string]UploadFunction
}

func (request *UploadRequest) ParseRequest(r *http.Request) error {
	request.Token = r.Header.Get("token")
	if len(request.Token) == 0 {
		log.Warn("Token is empty")
		return fmt.Errorf("token for upload wasn't provided")
	}
	request.Owner = db.TokenOwner(request.Token)
	if request.Owner == "" {
		log.Warn("Token is incorrect")
		return fmt.Errorf("incorrect token provided")
	}
	escapedPath := strings.Split(r.URL.EscapedPath(), "/")
	if len(escapedPath) < 4 {
		log.Warn("url for upload is not correct")
		return fmt.Errorf("incorrect Upload request")
	}
	request.Repo = escapedPath[3]
	r.ParseMultipartForm(1 << 31)
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Warn("Couldn't open file")
		return err
	}
	defer file.Close()
	request.Filename = header.Filename
	request.File = io.Reader(file) // multipart.sectionReadCloser
	limit := int64(db.QuotaLeft(request.Owner))
	if limit != -1 {
		request.File = io.LimitReader(file, limit)
	}
	log.Info(fmt.Sprintf("Printing io.Reader: %+v", request.File))
	if len(r.MultipartForm.Value["private"]) > 0 {
		request.Private = r.MultipartForm.Value["private"][0]
	}
	if len(r.MultipartForm.Value["version"]) > 0 {
		request.Version = r.MultipartForm.Value["version"][0]
	}
	if len(r.MultipartForm.Value["tags"]) > 0 {
		request.Tags = r.MultipartForm.Value["tags"][0]
	}
	return nil
}

type UploadFunction func() error

func (request *UploadRequest) InitUploaders() {
	request.uploaders = make(map[string]UploadFunction)
	request.uploaders["apt"] = request.UploadApt
	request.uploaders["raw"] = request.UploadRaw
	request.uploaders["template"] = request.UploadTemplate
}

func (request *UploadRequest) ExecRequest() error {
	uploader := request.uploaders[request.Repo]
	return uploader()
}

func (request *UploadRequest) BuildResult() *Result {
	result := new(Result)
	myUUID, _ := uuid.NewV4()
	request.fileID = myUUID.String()
	result.FileID = request.fileID
	result.Filename = request.Filename
	result.Repo = request.Repo
	result.Owner = request.Owner
	result.Tags = request.Tags
	result.Version = request.Version
	result.Md5 = request.md5
	result.Sha256 = request.sha256
	result.Size = request.size
	return result
}

func (request *UploadRequest) HandlePrivate() {
	if request.Private == "true" {
		db.MakePrivate(request.fileID, request.Owner)
	} else {
		db.MakePublic(request.fileID, request.Owner)
	}
}

func (request *UploadRequest) ReadDeb() (control bytes.Buffer, err error) {
	file, err := os.Open(config.Storage.Path + request.Filename)
	log.Check(log.WarnLevel, "Opening deb package", err)
	defer file.Close()
	library := ar.NewReader(file)
	for header, err := library.Next(); err != io.EOF; header, err = library.Next() {
		if err != nil {
			return control, err
		}
		if header.Name == "control.tar.gz" {
			ungzip, err := gzip.NewReader(library)
			if err != nil {
				return control, err
			}
			defer ungzip.Close()
			tr := tar.NewReader(ungzip)
			for tarHeader, err := tr.Next(); err != io.EOF; tarHeader, err = tr.Next() {
				if err != nil {
					return control, err
				}
				if tarHeader.Name == "./control" {
					if _, err := io.Copy(&control, tr); err != nil {
						return control, err
					}
					break
				}
			}
		}
	}
	return
}

func GetControl(control bytes.Buffer) map[string]string {
	d := make(map[string]string)
	for _, v := range strings.Split(control.String(), "\n") {
		line := strings.Split(v, ":")
		if len(line) > 1 {
			d[line[0]] = strings.TrimPrefix(line[1], " ")
		}
	}
	return d
}

func (request *UploadRequest) UploadApt() error {
	control, err := request.ReadDeb()
	if err != nil {
		return err
	}
	info := GetControl(control)
	result := request.BuildResult()
	result.Architecture = info["Architecture"]
	result.Description = info["Description"]
	result.Version = info["Version"]
	log.Info(fmt.Sprintf("Uploading apt file -> %+v", result))
	WriteDB(result)
	request.HandlePrivate()
	return nil
}

func (request *UploadRequest) UploadRaw() error {
	result := request.BuildResult()
	log.Info(fmt.Sprintf("Uploading raw file -> %+v", result))
	WriteDB(result)
	request.HandlePrivate()
	return nil
}

func LoadConfiguration(request *UploadRequest) (configuration string, err error) {
	var configurationFile bytes.Buffer
	f, err := os.Open(config.Storage.Path + request.md5)
	if err != nil {
		log.Warn("Failed to open template to load configuration file")
		return
	}
	defer f.Close()
	gzFile, err := gzip.NewReader(f)
	if err != nil {
		log.Warn("Failed to open .gz to load configuration file")
		return
	}
	tarFile := tar.NewReader(gzFile)
	for file, fileErr := tarFile.Next(); fileErr != io.EOF; file, err = tarFile.Next() {
		if file.Name == "config" {
			if _, err = io.Copy(&configurationFile, tarFile); err != nil {
				log.Warn("Failed to copy configuration file")
				return
			}
			break
		}
	}
	configuration = configurationFile.String()
	return
}

func FormatConfiguration(request *UploadRequest, configuration string) (template *Result) {
	template = request.BuildResult()
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
	searchRequest := new(SearchRequest)
	searchRequest.Name = template.Parent
	searchRequest.Repo = request.Repo
	searchRequest.Token = request.Token
	searchRequest.Operation = "list"
	list := searchRequest.Retrieve()
	for _, result := range list {
		if result.Name == template.Parent &&
			result.Owner == template.ParentOwner &&
			result.Version == template.ParentVersion {
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
	name, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", template.Name)
	version, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", template.Version)
	if name && version {
		return nil
	}
	return fmt.Errorf("name or version format is wrong")
}

func (request *UploadRequest) UploadTemplate() error {
	configuration, err := LoadConfiguration(request)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't upload template: %v", err))
		return err
	}
	result := FormatConfiguration(request, configuration)
	err = request.TemplateCheckValid(result)
	if err != nil {
		if err.Error() != "owner in config file is different" {
			db.QuotaUsageSet(request.Owner, -int(request.size))
			os.Remove(config.Storage.Path + request.md5)
		}
		log.Warn(fmt.Sprintf("Not valid template: %v", err))
		return err
	}
	request.fileID = result.FileID
	log.Info(fmt.Sprintf("Uploading template -> %+v", result))
	WriteDB(result)
	request.HandlePrivate()
	return nil
}

func (request *UploadRequest) Upload() error {
	filePath := config.Storage.Path + request.Filename
	log.Info(fmt.Sprintf("Uploading file: %s", filePath))
	file, err := os.Create(filePath)
	if err != nil {
		log.Warn("Couldn't create file for writing - %s", filePath)
		return err
	}
	defer file.Close()
	limit := int64(db.QuotaLeft(request.Owner))
	log.Info("file created for writing: ", filePath)
	log.Info("space left for user ", request.Owner, " is ", limit)
	log.Info(fmt.Sprintf("Copying file %+v to %+v", request.File, file))
	request.size, err = io.Copy(file, request.File)
	log.Info(fmt.Sprintf("request.size: %+v", request.size))
	if limit != -1 && (request.size == limit || err != nil) {
		log.Warn("User " + request.Owner + " exceeded storage quota, removing file")
		os.Remove(filePath)
		return fmt.Errorf("failed to write file or storage quota exceeded")
	} else {
		db.QuotaUsageSet(request.Owner, int(request.size))
		log.Info("User " + request.Owner + ", quota usage +" + strconv.Itoa(int(request.size)))
	}
	request.md5 = Hash(filePath, "md5")
	request.sha256 = Hash(filePath, "sha256")
	if len(request.md5) == 0 || len(request.sha256) == 0 {
		log.Warn("Failed to calculate hash for " + request.Filename)
		return fmt.Errorf("failed to calculate hash")
	}
	if request.Repo != "apt" {
		md5Path := config.Storage.Path + request.md5
		os.Rename(filePath, md5Path)
		log.Info(fmt.Sprintf("repo is not apt: renamed %s to %s", filePath, md5Path))
	}
	log.Info("First Upload stage completed")
	return request.ExecRequest()
}
