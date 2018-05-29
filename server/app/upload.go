package app

import (
	"io"
	"net/http"
	"fmt"
	"strings"
	"os"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"strconv"
	"crypto/md5"
	"crypto/sha512"
	"crypto/sha256"
	"crypto/sha1"
)

type UploadRequest struct {
	File     io.Reader
	Filename string
	Repo     string
	Owner    string
	Token    string
	Private  string
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

func Hash(file string, algo string) string {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to open file %s to calculate hash", file))
		return ""
	}
	hash := md5.New()
	switch algo {
	case "md5":
		hash = md5.New()
	case "sha1":
		hash = sha1.New()
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	}
	if _, err := io.Copy(hash, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

type uploadFunc func(request UploadRequest, md5, sha256 string) error

var (
	uploaders = make(map[string]uploadFunc)
)

func InitUploaders() {
	uploaders["apt"]      = UploadApt
	uploaders["raw"]      = UploadRaw
	uploaders["template"] = UploadTemplate
}

func (request UploadRequest) ExecRequest(md5, sha256 string) error {
	uploader := uploaders[request.Repo]
	return uploader(request, md5, sha256)
}

func UploadApt(request UploadRequest, md5, sha256 string) error {
	return nil
}

func UploadRaw(request UploadRequest, md5, sha256 string) error {
	return nil
}

func UploadTemplate(request UploadRequest, md5, sha256 string) error {
	return nil
}

func Upload(request UploadRequest) error {
	filePath := config.Storage.Path + request.Filename
	file, err := os.Create(filePath)
	if err != nil {
		log.Warn("Couldn't create file for writing - %s", filePath)
		return err
	}
	defer file.Close()
	limit := int64(db.QuotaLeft(request.Owner))
	if copied, err := io.Copy(file, request.File); limit != -1 && (copied == limit || err != nil) {
		log.Warn("User " + request.Owner + " exceeded storage quota, removing file")
		os.Remove(filePath)
		return fmt.Errorf("failed to write file or storage quota exceeded")
	} else {
		db.QuotaUsageSet(request.Owner, int(copied))
		log.Info("User " + request.Owner + ", quota usage +" + strconv.Itoa(int(copied)))
	}
	md5sum    := Hash(filePath, "md5")
	sha256sum := Hash(filePath, "sha256")
	if len(md5sum) == 0 || len(sha256sum) == 0 {
		log.Warn("Failed to calculate hash for " + request.Filename)
		return fmt.Errorf("failed to calculate hash")
	}
	if request.Repo != "apt" {
		md5Path := config.Storage.Path + md5sum
		os.Rename(filePath, md5Path)
		log.Debug(fmt.Sprintf("repo is not apt: renamed %s to %s", filePath, md5Path))
	}
	return request.ExecRequest(md5sum, sha256sum)
}
