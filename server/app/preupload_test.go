package app

import (
	"fmt"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/libgorjun"
	"os"
	"io"
	"path/filepath"
)

type PreUploadFunction func(int, gorjun.GorjunUser)

var (
	PreUploaders  map[int]PreUploadFunction
)

func InitPreUploaders() {
	PreUploaders = make(map[int]PreUploadFunction)
	PreUploaders[0] = PreUploadUnit
	PreUploaders[1] = PreUploadIntegration
}

func PrepareUploadRequest(scope int, user gorjun.GorjunUser, repo string, index int) (*UploadRequest) {
	file := Files[scope][user.Username][NamesLayer][index]
	filePath, _ := os.Open(Dirs[scope][user.Username] + file)
	request := &UploadRequest{
		File:     io.Reader(filePath),
		Filename: filepath.Base(file),
		Repo:     repo,
		Owner:    user.Username,
		Token:    user.Token,
		Private:  ScopeType(scope),
		Tags:     repo,
	}
//	filePath.Close()
	request.InitUploaders()
	return request
}

func PreUploadFileUnit(scope int, user gorjun.GorjunUser, repo string, index int) (*UploadRequest, error) {
	request := PrepareUploadRequest(scope, user, repo, index)
	err := request.Upload()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func PreUploadUnit(scope int, user gorjun.GorjunUser) {
	fileIDs := make([]string, 0)
	repos := []string{
		"raw", "template", "apt",
		"template", "template", "template", "template", "template", "template", "template",
	}
	for i := 0; i < len(Files[scope][user.Username][IDsLayer]); i++ {
		file := Files[scope][user.Username][NamesLayer][i]
		path := FilesDir + file
		repo := repos[i]
		request, err := PreUploadFileUnit(scope, user, repo, i)
		if err != nil {
			log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s: %v", path, repo, user.Username, user.Token, err)
		} else {
			fileIDs = append(fileIDs, request.fileID)
		}
		log.Info(fmt.Sprintf("Uploading %s file %s of user %s to repo %s finished", FileType(scope), path, user.Username, repo))
	}
	log.Info(fmt.Sprintf("%s files of user %s are pre-uploaded to CDN", FileType(scope), user.Username))
	UserFiles[scope][user.Username] = fileIDs
}

func PreUploadIntegration(scope int, user gorjun.GorjunUser) {
	fileIDs := make([]string, 0)
	dir := Dirs[scope][user.Username]
	repos := []string{
		"raw", "template", "apt",
		"template", "template", "template", "template", "template", "template", "template",
	}
	for i := 0; i < len(Files[scope][user.Username][IDsLayer]); i++ {
		filename := Files[scope][user.Username][NamesLayer][i]
		path := dir + filename
		repo := repos[i]
		fileID, err := user.Upload(path, repo, ScopeType(scope))
		if err != nil {
			log.Warn("Failed to upload %s, repo: %s, user: %s, token: %s", path, repo, user.Username, user.Token)
		} else {
			fileIDs = append(fileIDs, fileID)
		}
		log.Info(fmt.Sprintf("Uploading %s file %s of user %s to repo %s finished", FileType(scope), path, user.Username, repo))
	}
	log.Info(fmt.Sprintf("%s files of user %s are pre-uploaded to CDN", FileType(scope), user.Username))
	UserFiles[scope][user.Username] = fileIDs
}

func PreUploadAllFiles(user gorjun.GorjunUser) {
	PreUploaders[Integration](PublicScope, user)
	PreUploaders[Integration](PrivateScope, user)
	log.Info(fmt.Sprintf("All uploaded files of user %s: %+v", user.Username, UserFiles))
	return
}

func PreUpload() {
	log.Info("Pre-uploading files to CDN")
	PreUploadAllFiles(Subutai)
	PreUploadAllFiles(Lorem)
	PreUploadAllFiles(Ipsum)
	log.Info("Pre-uploading files finished")
}
