package app

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"
	"fmt"
	"os"
	"mime/multipart"
	"strings"
	"path/filepath"
	"github.com/subutai-io/agent/log"
	"strconv"
)

func PrepareRequest(token, filePath, repo, version, tags, private string) *http.Request {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	if token != "" {
		formWriter, _ := writer.CreateFormField("token")
		formWriter.Write([]byte(token))
	}
	if strings.Contains(filePath, "nothing") {
		// do not create file in request
	} else {
		formWriter, _ := writer.CreateFormFile("file", filepath.Base(filePath))
		file, _ := os.Create(filePath)
		io.Copy(formWriter, io.Reader(file))
//		file.Close()
	}
	if version != "" {
		formWriter, _ := writer.CreateFormField("version")
		formWriter.Write([]byte(version))
	}
	if tags != "" {
		formWriter, _ := writer.CreateFormField("tags")
		formWriter.Write([]byte(tags))
	}
	if private != "" {
		formWriter, _ := writer.CreateFormField("private")
		formWriter.Write([]byte(private))
	}
	writer.Close()
	request, _ := http.NewRequest("POST", fmt.Sprintf(Localhost + "/kurjun/rest/%s/upload", repo), &buffer)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("token", token)
	return request
}

func TestUploadRequest_ParseRequest(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "TestUploadRequest_ParseRequest-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 8; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	tests[0].args.r = PrepareRequest(Subutai.Token, Dirs[PublicScope][Subutai.Username] + "auxiliaryFile-0", "apt", "", "", "false")
	tests[0].wantErr = false
	tests[1].args.r = PrepareRequest(Subutai.Token, Dirs[PublicScope][Subutai.Username] + "auxiliaryFile-1", "raw", "", "Run", "true")
	tests[1].wantErr = false
	tests[2].args.r = PrepareRequest(Subutai.Token, Dirs[PublicScope][Subutai.Username] + "auxiliaryFile-2", "template", "7.0.0", "", "false")
	tests[2].wantErr = false
	tests[3].args.r = PrepareRequest(Subutai.Token, Dirs[PublicScope][Subutai.Username] + "auxiliaryFile-3", "apt", "7.0.0", "Sail", "true")
	tests[3].wantErr = false
	tests[4].args.r = PrepareRequest(Subutai.Token, Dirs[PublicScope][Subutai.Username] + "auxiliaryFile-4", "raw", "7.0.0", "Run,Sail", "false")
	tests[4].wantErr = false
	tests[5].args.r = PrepareRequest("", Dirs[PublicScope][Subutai.Username] + "auxiliaryFile-5", "template", "7.0.0", "Park", "true")
	tests[5].wantErr = true
	tests[6].args.r = PrepareRequest(Lorem.Token, Dirs[PublicScope][Lorem.Username] + "auxiliaryFile-6", "apt", "2.2.3", "nobodyreadstags", "false")
	tests[6].wantErr = false
	tests[7].args.r = PrepareRequest("incorrectToken", Dirs[PublicScope][Lorem.Username] + "auxiliaryFile-7-nothing", "raw", "5.0.2", "whoreadstagsanyway,nothing", "true")
	tests[7].wantErr = true
	tests[8].args.r = PrepareRequest(Lorem.Token, Dirs[PublicScope][Lorem.Username] + "auxiliaryFile-8-nothing", "template", "3.1.2", "unitTest", "false")
	tests[8].wantErr = true
	for _, tt := range tests {
		errored := false
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.ParseRequest(tt.args.r); (err != nil) != tt.wantErr {
				errored = true
				t.Errorf("UploadRequest.ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if errored {
			break
		}
	}
}

func TestUploadRequest_InitUploaders(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "TestUploadRequest_InitUploaders-"},
		// TODO: Add test cases.
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			request.InitUploaders()
			if len(request.uploaders) == 3 {
				log.Info("OK")
			} else {
				t.Errorf("uploaders uninitialized")
			}
		})
	}
}

func TestUploadRequest_ExecRequest(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "TestUploadRequest_ExecRequest-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 16; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	{
		repos := []string{"raw", "apt", "template", "raw"}
		privates := []string{"false", "false", "false", "true"}
		wantErrs := []bool{false, true, true, false}
		for i := 0; i <= 3; i++ {
			testNumber := strconv.Itoa(i)
			file := "auxiliaryFile-" + testNumber
			filePath, _ := os.Open(Dirs[PublicScope][Lorem.Username] + file)
			path := FilesDir + file
			auxiliaryFile, _ := os.Create(path)
			io.Copy(auxiliaryFile, io.Reader(filePath))
			auxiliaryFile.Close()
			auxiliaryFileStats, _ := os.Stat(path)
			md5Sum := Hash(path, "md5")
			sha256Sum := Hash(path, "sha256")
			os.Rename(path, FilesDir + md5Sum)
			tests[i].fields = fields {
				File:     io.Reader(filePath),
				Filename: filepath.Base(file),
				Repo:     repos[i],
				Owner:    Lorem.Username,
				Token:    Lorem.Token,
				Private:  privates[i],
				Tags:     "tag-" + testNumber,
				Version:  "7.0." + testNumber,
				md5:      md5Sum,
				sha256:   sha256Sum,
				size:     auxiliaryFileStats.Size(),
			}
//			filePath.Close()
			tests[i].wantErr = wantErrs[i]
		}
	}
	{
		repos := []string{"raw", "template"}
		users := []gorjun.GorjunUser{Ipsum, Lorem, Subutai}
		for i := 4; i <= 15; i++ {
			user := users[(i - 4) / 2 % 3]
			scope := (i - 4) / 6
			repo := repos[i & 1]
			testNumber := strconv.Itoa(i)
			request := PrepareUploadRequest(scope, user, repo, i & 1)
			file := Files[scope][user.Username][NamesLayer][i & 1]
			path := FilesDir + file
			auxiliaryFile, _ := os.Create(path)
			io.Copy(auxiliaryFile, request.File)
			auxiliaryFile.Close()
			auxiliaryFileStats, _ := os.Stat(path)
			md5Sum := Hash(path, "md5")
			sha256Sum := Hash(path, "sha256")
			os.Rename(path, FilesDir + md5Sum)
			version := ""
			if (i & 1) == 0 {
				version = "7.0." + testNumber
			}
			tests[i].fields = fields {
				File:     request.File,
				Filename: request.Filename,
				Repo:     request.Repo,
				Owner:    request.Owner,
				Token:    request.Token,
				Private:  request.Private,
				Tags:     "tag-" + testNumber,
				Version:  version,
				md5:      md5Sum,
				sha256:   sha256Sum,
				size:     auxiliaryFileStats.Size(),
			}
			tests[i].wantErr = false
		}
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "apt", 2)
		file := Files[PublicScope][Lorem.Username][NamesLayer][2]
		path := FilesDir + file
		auxiliaryFile, _ := os.Create(path)
		io.Copy(auxiliaryFile, request.File)
		auxiliaryFile.Close()
		auxiliaryFileStats, _ := os.Stat(path)
		md5Sum := Hash(path, "md5")
		sha256Sum := Hash(path, "sha256")
		tests[16].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "apt",
			Owner:    request.Owner,
			Token:    request.Token,
			Private:  request.Private,
			Tags:     "tag-16",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxiliaryFileStats.Size(),
		}
		tests[16].wantErr = false
	}
	for _, tt := range tests {
		errored := false
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			request.InitUploaders()
			if err := request.ExecRequest(); (err != nil) != tt.wantErr {
				errored = true
				t.Errorf("UploadRequest.ExecRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if errored {
			break
		}
	}
}

func TestUploadRequest_BuildResult(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name   string
		fields fields
		want   *Result
	}{
		{name: "TestUploadRequest_BuildResult-"},
		// TODO: Add test cases.
	}
	tests[0].name += "0"
	tests[0].fields = fields{
		Filename: "some-filename",
		Repo:     "some-repo",
		Owner:    "some-owner",
		Tags:     "some-tags,some-tag",
		Version:  "some-version",
		md5:      "some-md5",
		sha256:   "some-sha256",
		size:     123,
	}
	tests[0].want = &Result{
		Filename: "some-filename",
		Repo:     "some-repo",
		Owner:    "some-owner",
		Tags:     "some-tags,some-tag",
		Version:  "some-version",
		Md5:      "some-md5",
		Sha256:   "some-sha256",
		Size:     123,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			got := request.BuildResult()
			tt.want.FileID = got.FileID
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UploadRequest.BuildResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadRequest_HandlePrivate(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "TestUploadRequest_HandlePrivate-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 1; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	tests[0].fields = fields{
		fileID:   tests[0].name,
		Owner:    Subutai.Username,
		Filename: tests[0].name,
		Private:  "false",
	}
	tests[1].fields = fields{
		fileID:   tests[1].name,
		Owner:    Lorem.Username,
		Filename: tests[1].name,
		Private:  "true",
	}
	for _, tt := range tests {
		errored := false
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			result := request.BuildResult()
			result.FileID = tt.fields.fileID
			request.fileID = tt.fields.fileID
			result.Repo = "raw"
			log.Info(fmt.Sprintf("Going to write result: %+v, request: %+v", result, request))
//			WriteDB(result)
			request.HandlePrivate()
			if !DB.CheckShare(tt.fields.fileID, tt.fields.Owner) {
				errored = true
				t.Errorf("file is not available to its owner")
			}
		})
		if errored {
			break
		}
	}
}

func TestUploadRequest_ReadDeb(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name        string
		fields      fields
		wantControl bytes.Buffer
		wantErr     bool
	}{
		{name: "TestUploadRequest_ReadDeb-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 1; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	{
		file := Files[PublicScope][Lorem.Username][NamesLayer][2]
		filePath, _ := os.Open(Dirs[PublicScope][Lorem.Username] + file)
		path := FilesDir + file
		auxiliaryFile, _ := os.Create(path)
		io.Copy(auxiliaryFile, io.Reader(filePath))
		auxiliaryFile.Close()
		auxiliaryFileStats, _ := os.Stat(path)
		md5Sum := Hash(path, "md5")
		sha256Sum := Hash(path, "sha256")
		tests[0].fields = fields{
			File:     io.Reader(filePath),
			Filename: filepath.Base(file),
			Repo:     "apt",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  "false",
			Tags:     "tag-0",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxiliaryFileStats.Size(),
		}
//		filePath.Close()
		tests[0].wantErr = false
	}
	tests[1].wantErr = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
//			gotControl, err := request.ReadDeb()
			_, err := request.ReadDeb()
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.ReadDeb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
//			if !reflect.DeepEqual(gotControl, tt.wantControl) {
//				t.Errorf("UploadRequest.ReadDeb() = %v, want %v", gotControl, tt.wantControl)
//			}
		})
	}
}

func TestGetControl(t *testing.T) {
	Integration = 0
	type args struct {
		control bytes.Buffer
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetControl(tt.args.control); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetControl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadRequest_UploadApt(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.UploadApt(); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.UploadApt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_UploadRaw(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.UploadRaw(); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.UploadRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfiguration(t *testing.T) {
	Integration = 0
	type args struct {
		request *UploadRequest
	}
	tests := []struct {
		name              string
		args              args
		wantConfiguration string
		wantErr           bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfiguration, err := LoadConfiguration(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotConfiguration != tt.wantConfiguration {
				t.Errorf("LoadConfiguration() = %v, want %v", gotConfiguration, tt.wantConfiguration)
			}
		})
	}
}

func TestFormatConfiguration(t *testing.T) {
	Integration = 0
	type args struct {
		request       *UploadRequest
		configuration string
	}
	tests := []struct {
		name         string
		args         args
		wantTemplate *Result
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTemplate := FormatConfiguration(tt.args.request, tt.args.configuration); !reflect.DeepEqual(gotTemplate, tt.wantTemplate) {
				t.Errorf("FormatConfiguration() = %v, want %v", gotTemplate, tt.wantTemplate)
			}
		})
	}
}

func TestUploadRequest_TemplateCheckValid(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	Integration = 0
	type args struct {
		template *Result
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.TemplateCheckValid(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_TemplateCheckFieldsPresent(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	type args struct {
		template *Result
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.TemplateCheckFieldsPresent(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckFieldsPresent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_TemplateCheckOwner(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	type args struct {
		template *Result
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.TemplateCheckOwner(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckOwner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_TemplateCheckDependencies(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	type args struct {
		template *Result
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.TemplateCheckDependencies(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_TemplateCheckFormat(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	type args struct {
		template *Result
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "TestUploadRequest_TemplateCheckFormat-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 3; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	names := []string{"name", "!"}
	versions:= []string{"7.0.0", "!"}
	for i := 0; i < 4; i++ {
		tests[i].args.template = &Result{
			Name: names[i & 1],
			Version: versions[(i >> 1) & 1],
		}
		if i > 0 {
			tests[i].wantErr = true
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.TemplateCheckFormat(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_UploadTemplate(t *testing.T) {
	Integration = 0
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			if err := request.UploadTemplate(); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.UploadTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_Upload(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	PreDownload()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "TestUploadRequest_Upload-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 25; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	{
		repos := []string{"raw", "apt", "template", "raw"}
		privates := []string{"false", "false", "false", "true"}
		wantErrs := []bool{false, true, true, false}
		for i := 0; i <= 3; i++ {
			testNumber := strconv.Itoa(i)
			file := "auxiliaryFile-" + testNumber
			filePath, _ := os.Create(Dirs[PublicScope][Lorem.Username] + file)
			tests[i].fields = fields {
				File:     io.Reader(filePath),
				Filename: filepath.Base(file),
				Repo:     repos[i],
				Owner:    Lorem.Username,
				Token:    Lorem.Token,
				Private:  privates[i],
				Tags:     "tag-" + testNumber,
				Version:  "7.0." + testNumber,
			}
//			filePath.Close()
			tests[i].wantErr = wantErrs[i]
		}
	}
	{
		repos := []string{"raw", "template"}
		users := []gorjun.GorjunUser{Ipsum, Lorem, Subutai}
		for i := 4; i <= 15; i++ {
			user := users[(i - 4) / 2 % 3]
			scope := (i - 4) / 6
			repo := repos[i & 1]
			testNumber := strconv.Itoa(i)
			request := PrepareUploadRequest(scope, user, repo, i & 1)
			version := ""
			if (i & 1) == 0 {
				version = "7.0." + testNumber
			}
			log.Info(fmt.Sprintf("request.File: %+v", request.File))
			tests[i].fields = fields {
				File:     request.File,
				Filename: filepath.Base(request.Filename),
				Repo:     request.Repo,
				Owner:    request.Owner,
				Token:    request.Token,
				Private:  request.Private,
				Tags:     "tag-" + testNumber,
				Version:  version,
			}
			tests[i].wantErr = false
		}
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "apt", 2)
		tests[16].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "apt",
			Owner:    request.Owner,
			Token:    request.Token,
			Private:  request.Private,
			Tags:     "tag-16",
		}
		tests[16].wantErr = false
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 1)
		tests[17].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Ipsum.Username,
			Token:    Ipsum.Token,
			Private:  request.Private,
			Tags:     "tag-17",
		}
		tests[17].wantErr = true
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 3)
		tests[18].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Ipsum.Username,
			Token:    Ipsum.Token,
			Private:  request.Private,
			Tags:     "tag-18",
		}
		tests[18].wantErr = true
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 0)
		tests[19].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Ipsum.Username,
			Token:    Ipsum.Token,
			Private:  request.Private,
			Tags:     "tag-19",
		}
		tests[19].wantErr = true
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 4)
		tests[20].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  request.Private,
			Tags:     "tag-20",
		}
		tests[20].wantErr = false
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 5)
		tests[21].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  request.Private,
			Tags:     "tag-21",
		}
		tests[21].wantErr = true
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 6)
		tests[22].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  request.Private,
			Tags:     "tag-22",
		}
		tests[22].wantErr = true
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 7)
		tests[23].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  request.Private,
			Tags:     "tag-23",
		}
		tests[23].wantErr = false
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 8)
		tests[24].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  request.Private,
			Tags:     "tag-24",
		}
		tests[24].wantErr = true
	}
	{
		request := PrepareUploadRequest(PublicScope, Lorem, "template", 9)
		tests[25].fields = fields {
			File:     request.File,
			Filename: request.Filename,
			Repo:     "template",
			Owner:    Lorem.Username,
			Token:    Lorem.Token,
			Private:  request.Private,
			Tags:     "tag-25",
		}
		tests[25].wantErr = true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &UploadRequest{
				File:      tt.fields.File,
				Filename:  tt.fields.Filename,
				Repo:      tt.fields.Repo,
				Owner:     tt.fields.Owner,
				Token:     tt.fields.Token,
				Private:   tt.fields.Private,
				Tags:      tt.fields.Tags,
				Version:   tt.fields.Version,
				fileID:    tt.fields.fileID,
				md5:       tt.fields.md5,
				sha256:    tt.fields.sha256,
				size:      tt.fields.size,
				uploaders: tt.fields.uploaders,
			}
			request.InitUploaders()
			if err := request.Upload(); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.Upload() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				log.Info(fmt.Sprintf("%s correct: %v", tt.name, err))
			}
		})
	}
}

func TestUploadRequest_UploadStress(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "TestUploadRequest_UploadStress-"},
		// TODO: Add test cases.
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}

/*
func TestUploadRequest_UploadStress(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type fields struct {
		File      io.Reader
		Filename  string
		Repo      string
		Owner     string
		Token     string
		Private   string
		Tags      string
		Version   string
		fileID    string
		md5       string
		sha256    string
		size      int64
		uploaders map[string]UploadFunction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "TestUploadRequest_UploadStress-"},
		// TODO: Add test cases.
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := 0
			stop := false
			for !stop {
				finished := make(chan bool, 100)
				j := i
				for k := 0; k < 100; k++ {
					go func() {
						j++
						file := Dirs[PublicScope][Lorem.Username] + strconv.Itoa(j) + ".txt"
						log.Info(fmt.Sprintf("File: %+v", file))
						ioutil.WriteFile(file, []byte(file), 0755)
//						cmd := exec.Command("wget", "-O", file, Raw + Files[PublicScope][Lorem.Username][IDsLayer][4])
//						cmd.Run()
						filePath, _ := os.Open(file)
						request := &UploadRequest{
							File:     io.Reader(filePath),
							Filename: filepath.Base(file),
							Repo:     "raw",
							Owner:    Lorem.Username,
							Token:    Lorem.Token,
							Private:  ScopeType(PublicScope),
							Tags:     "tag-" + strconv.Itoa(j),
						}
						tt.wantErr = true
						request.InitUploaders()
						if err := request.Upload(); err != nil {
							stop = true
							log.Info(fmt.Sprintf("Stress test successfull"))
						} else {
							log.Warn(fmt.Sprintf("Still not enough files, starting next attempt"))
						}
						finished <- true
					}()
				}
				for j := 0; j < 100; j++ {
					<-finished
				}
				close(finished)
				i += 100
			}
		})
	}
}
*/