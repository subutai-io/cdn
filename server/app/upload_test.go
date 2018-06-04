package app

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"
	"github.com/subutai-io/cdn/db"
	"fmt"
	"os"
	"mime/multipart"
	"strings"
	"path/filepath"
	"github.com/subutai-io/agent/log"
	"os/exec"
)

var (
	SubutaiName = "subutai"
	SubutaiKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
=R8TO
-----END PGP PUBLIC KEY BLOCK-----
`
	SubutaiToken = "15a5237ee5314282e52156cfad72e86b53ef0ad47baecc31233dbb1c06f4327c"
	AkenzhalievName = "akenzhaliev"
	AkenzhalievKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
iTL3bcFog5Zf+w==
=qbAq
-----END PGP PUBLIC KEY BLOCK-----
`
	AkenzhalievToken = "7327753f12b67d440f481c27e461925513d30cf2d56b6ac16060aad1021c293d"
	AbaytulakovaName = "abaytulakova"
	AbaytulakovaKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mQENBFr5Z7wBCACv7AytUuenUbXTs8jd9Pvd3j832vWGHwX9Kf4v6iW4INftGguQ
0Q40nvYKKhFmaeBUNkm3p13iXqq8/rjTFuC+kkDaT2AP72UsQ3KSF6K15cKC8gkF
ktyNZrAraofQ+n4rCGSpHKldf+8bU7+mpnGEEGKsfomOTzWwAsmoilGAseFr5tV9
x5YJIrzcMoiymCqsiOgLTq7vAYPQOl5gcFxlMSt7/cgpWHjTqBCUtoJ6Kvt8mftE
Pd1ycK6mEoQ0qROuazpQCtMeeC5IVFfw27pR1Jrpf6dbMJCqgt9FUIf0aDdrxr+l
TazFSUrTOHYooSP6bDaRjVQ/k5IZnC43M/3lABEBAAG0MGFiYXl0dWxha292YSA8
YWJheXR1bGFrb3ZhQG9wdGltYWwtZHluYW1pY3MuY29tPokBOAQTAQIAIgUCWvln
vAIbAwYLCQgHAwIGFQgCCQoLBBYCAwECHgECF4AACgkQ5L9rtPCyVCeffgf/R0Wz
nIhaji5iwkDHlPR3FyENlWD6KUwj8DhD8t5cTmlPuLVBf7x92TTPJpoqOVtXObgI
yXPPS++49Paubh5F9kIPOca2ilmHReum7XzdR2pAL/EVnkZn37XgW3JN21F645iE
ZmUA/INhpxpXDpRVeF1ZwQ+5/VghXLkZJ4BOwCUwjrtZxb/FhHbCmbrY3Sq1FWZ9
45btKOrhrmlRpB4i3bM+hNZfeqzVpQ1SvboCqJ1r+ZOvUwcZ+RFPUk/z4xiG1d7v
8D3Q66G+6pXmLlAs4/9lQ9wlGMucvChsXIg07qnNkAgaYGh65tI+oEjKDXGPCPr2
bPLPCEwvkIfJivKKybkBDQRa+We8AQgA2AqyFsETnEEOxSbZFphqdRMC8xlEekmj
P9TMZDjrMouRKTJs8Hx/AU9kuKubua0hJWYft8QMlR3JjFkPKhRGs8LkEra7mGSK
6XzKarH5oaF2toFUxIzE6un922wGDt2kq4kiTA0+SI2QX15ZvOX/3+ZYR/q0DV+2
wpsBX2UR3y3UjHIk5BjMU+E9wJNw86098nZ6l8b10xZzwf+rSxpDVwB+1lJTAFbB
ivtoTKxrxmaRKvh909ut6TJA6uU5tGL1AJ2ATysuOLQEmsASnRikxomjt6CMjQGG
Lx3NknQ26BmYfh8rrDQknWw1E0i5mLUgaJJNKv1sRNd4fwhSYAGiWQARAQABiQEf
BBgBAgAJBQJa+We8AhsMAAoJEOS/a7TwslQn3voH/0y0aXSxvtt0gd62+6iA5R6m
Y7bQvf4vyndEXk4+Z3LjmMRTfoxak/AhnxzAkkC2ednVBscZE9btxiP7dkdHlJ4B
lvbR5K7xRJyFgUFUbL+s5GyXh9j0OVruw6qFteX04UD7OTZJxTm/+yIzLwmDfovP
NkfUgyMjtjPjrKWyydoEbofg2WTFgiuA6f/5zHZLG6NNxTgoxz6ZSLhnvUN3JVJ5
YaQ9+dgz72Zum+hRgg7t7cN4VF8iMuK2fYx/grIZ2wH5c4s9wJsPy0D+MiijhaF5
fOBXzFkOhrc9NzsD4RGldnFJQkaYotxOXbPlfosCdc+3WuS0Np+mpnu6/LsfzSQ=
=XRVg
-----END PGP PUBLIC KEY BLOCK-----
`
	AbaytulakovaToken = "37fc4c3ef862c079ea44da0b7863948d10e8493c514d83e751a6253363f564cf"
)

func PrepareUsersAndTokens() {
	db.RegisterUser([]byte(SubutaiName), []byte(SubutaiKey))
	db.RegisterUser([]byte(AkenzhalievName), []byte(AkenzhalievKey))
	db.RegisterUser([]byte(AbaytulakovaName), []byte(AbaytulakovaKey))
	db.SaveToken(SubutaiName, SubutaiToken)
	db.SaveToken(AkenzhalievName, AkenzhalievToken)
	db.SaveToken(AbaytulakovaName, AbaytulakovaToken)
}

func PrepareRequest(token, filePath, repo, version, tags, private string) *http.Request {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)

	if token != "" {
		fw, _ := w.CreateFormField("token")
		fw.Write([]byte(token))
	}

	if strings.Contains(filePath, "nothing") {
		// nothing :)
	} else {
		fw, _ := w.CreateFormFile("file", filepath.Base(filePath))
		f, _ := os.Create(filePath)
		io.Copy(fw, f)
		f.Close()
	}

	if version != "" {
		fw, _ := w.CreateFormField("version")
		fw.Write([]byte(version))
	}

	if tags != "" {
		fw, _ := w.CreateFormField("tags")
		fw.Write([]byte(tags))
	}

	if private != "" {
		fw, _ := w.CreateFormField("private")
		fw.Write([]byte(private))
	}

	w.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:8080/kurjun/rest/%s/upload", repo), &b)

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("token", token)

	return req
}

func TestUploadRequest_ParseRequest(t *testing.T) {
	SetUp(false)
	PrepareUsersAndTokens()
	defer TearDown(false)
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
		{name: "TestUploadRequest_ParseRequest-0"},
		{name: "TestUploadRequest_ParseRequest-1"},
		{name: "TestUploadRequest_ParseRequest-2"},
		{name: "TestUploadRequest_ParseRequest-3"},
		{name: "TestUploadRequest_ParseRequest-4"},
		{name: "TestUploadRequest_ParseRequest-5"},
		{name: "TestUploadRequest_ParseRequest-6"},
		{name: "TestUploadRequest_ParseRequest-7"},
		{name: "TestUploadRequest_ParseRequest-8"},
		// TODO: Add test cases.
	}
	tests[0].args.r = PrepareRequest(SubutaiToken, "/tmp/data/public/subutai/AWOLNATION", "apt", "", "", "false")
	tests[0].wantErr = false
	tests[1].args.r = PrepareRequest(SubutaiToken, "/tmp/data/public/subutai/AWOLNATION-Run", "raw", "", "Run", "true")
	tests[1].wantErr = false
	tests[2].args.r = PrepareRequest(SubutaiToken, "/tmp/data/public/subutai/AWOLNATION", "template", "7.0.0", "", "false")
	tests[2].wantErr = false
	tests[3].args.r = PrepareRequest(SubutaiToken, "/tmp/data/public/subutai/AWOLNATION-Sail", "apt", "7.0.0", "Sail", "true")
	tests[3].wantErr = false
	tests[4].args.r = PrepareRequest(SubutaiToken, "/tmp/data/public/subutai/AWOLNATION-Run-Sail", "raw", "7.0.0", "Run,Sail", "false")
	tests[4].wantErr = false
	tests[5].args.r = PrepareRequest("", "/tmp/data/public/subutai/Linkin", "template", "7.0.0", "Park", "true")
	tests[5].wantErr = true
	tests[6].args.r = PrepareRequest(AkenzhalievToken, "/tmp/data/public/akenzhaliev/Park", "apt", "2.2.3", "nobodyreadstags", "false")
	tests[6].wantErr = false
	tests[7].args.r = PrepareRequest("incorrectToken", "/tmp/data/public/akenzhaliev/nothing", "raw", "5.0.2", "whoreadstagsanyway,nothing", "true")
	tests[7].wantErr = true
	tests[8].args.r = PrepareRequest(AkenzhalievToken, "/tmp/data/public/akenzhaliev/nothing", "template", "3.1.2", "unitTest", "false")
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
		{name: "TestUploadRequest_InitUploaders-0"},
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
			request.InitUploaders()
			if len(request.uploaders) == 3 {
				log.Info("OK")
			} else {
				t.Errorf("uploaders uninitialized")
			}
//			apt := UploadFunction(request.UploadApt)
//			raw := UploadFunction(request.UploadRaw)
//			template := UploadFunction(request.UploadTemplate)
//			aptPointer := &apt
//			rawPointer := &raw
//			templatePointer := &template
//			if len(request.uploaders) == 3 {
//				uploaderApt := request.uploaders["apt"]
//				uploaderAptPointer := &uploaderApt
//				uploaderRaw := request.uploaders["raw"]
//				uploaderRawPointer := &uploaderRaw
//				uploaderTemplate := request.uploaders["template"]
//				uploaderTemplatePointer := &uploaderTemplate
//				if aptPointer == uploaderAptPointer &&
//					rawPointer == uploaderRawPointer &&
//					templatePointer == uploaderTemplatePointer {
//					log.Info("OK")
//				} else {
//					t.Errorf("%s failed: \"uploaders\" uses unexpected functions")
//				}
//			} else {
//				t.Errorf("%s failed: not all uploaders initialized", tt.name)
//			}
		})
	}
}

func TestUploadRequest_ExecRequest(t *testing.T) {
	SetUp(false)
	PrepareUsersAndTokens()
	defer TearDown(false)
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
		{name: "TestUploadRequest_ExecRequest-0"},
		{name: "TestUploadRequest_ExecRequest-1"},
		{name: "TestUploadRequest_ExecRequest-2"},
		{name: "TestUploadRequest_ExecRequest-3"},
		{name: "TestUploadRequest_ExecRequest-4"},
		{name: "TestUploadRequest_ExecRequest-5"},
		{name: "TestUploadRequest_ExecRequest-6"},
		{name: "TestUploadRequest_ExecRequest-7"},
		{name: "TestUploadRequest_ExecRequest-8"},
		{name: "TestUploadRequest_ExecRequest-9"},
		// TODO: Add test cases.
	}
	{
		auxFile, _ := os.Create("/tmp/data/files/aux-0")
		auxFileStats, _ := os.Stat("/tmp/data/files/aux-0")
		md5Sum := Hash("/tmp/data/files/aux-0", "md5")
		sha256Sum := Hash("/tmp/data/files/aux-0", "sha256")
		os.Rename("/tmp/data/files/aux-0", "/tmp/data/files/" + md5Sum)
		tests[0].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "aux-0",
			Repo:     "raw",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "false",
			Tags:     "",
			Version:  "",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[0].wantErr = false
	}
	{
		auxFile, _ := os.Create("/tmp/data/files/aux-1")
		auxFileStats, _ := os.Stat("/tmp/data/files/aux-1")
		md5Sum := Hash("/tmp/data/files/aux-1", "md5")
		sha256Sum := Hash("/tmp/data/files/aux-1", "sha256")
		tests[1].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "aux-1",
			Repo:     "apt",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "true",
			Tags:     "",
			Version:  "",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[1].wantErr = true
	}
	{
		auxFile, _ := os.Create("/tmp/data/files/aux-2")
		auxFileStats, _ := os.Stat("/tmp/data/files/aux-2")
		md5Sum := Hash("/tmp/data/files/aux-2", "md5")
		sha256Sum := Hash("/tmp/data/files/aux-2", "sha256")
		os.Rename("/tmp/data/files/aux-2", "/tmp/data/files/" + md5Sum)
		tests[2].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "aux-2",
			Repo:     "template",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "false",
			Tags:     "",
			Version:  "",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[2].wantErr = true
	}
	{
		auxFile, _ := os.Create("/tmp/data/files/aux-3")
		auxFileStats, _ := os.Stat("/tmp/data/files/aux-3")
		md5Sum := Hash("/tmp/data/files/aux-3", "md5")
		sha256Sum := Hash("/tmp/data/files/aux-3", "sha256")
		os.Rename("/tmp/data/files/aux-3", "/tmp/data/files/" + md5Sum)
		tests[3].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "aux-3",
			Repo:     "raw",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "true",
			Tags:     "nobodyreadstags",
			Version:  "7.0.0",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[3].wantErr = false
	}
	{
		cmd := exec.Command("wget", "-O", "/tmp/data/public/abaytulakova/influxdb_1.2.2_amd64.deb", "https://cdn.subutai.io:8338/kurjun/rest/apt/influxdb_1.2.2_amd64.deb")
		cmd.Run()
		log.Info("Downloaded file")
		actualFile, _ := os.Open("/tmp/data/public/abaytulakova/influxdb_1.2.2_amd64.deb")
		auxFile, _ := os.Create("/tmp/data/files/influxdb_1.2.2_amd64.deb")
		io.Copy(auxFile, actualFile)
		auxFileStats, _ := os.Stat("/tmp/data/files/influxdb_1.2.2_amd64.deb")
		md5Sum := Hash("/tmp/data/files/influxdb_1.2.2_amd64.deb", "md5")
		sha256Sum := Hash("/tmp/data/files/influxdb_1.2.2_amd64.deb", "sha256")
		tests[4].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "influxdb_1.2.2_amd64.deb",
			Repo:     "apt",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "false",
			Tags:     "nobodyreadstags",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[4].wantErr = false
	}
	{
		cmd := exec.Command("wget", "-O", "/tmp/data/public/subutai/debian-stretch-subutai-template_0.4.1_amd64.tar.gz", "https://cdn.subutai.io:8338/kurjun/rest/template/download?id=50778fae-f028-42ab-beff-437e4f34ee48")
		cmd.Run()
		log.Info("Downloaded file")
		actualFile, _ := os.Open("/tmp/data/public/subutai/debian-stretch-subutai-template_0.4.1_amd64.tar.gz")
		auxFile, _ := os.Create("/tmp/data/files/debian-stretch-subutai-template_0.4.1_amd64.tar.gz")
		io.Copy(auxFile, actualFile)
		auxFileStats, _ := os.Stat("/tmp/data/files/debian-stretch-subutai-template_0.4.1_amd64.tar.gz")
		md5Sum := Hash("/tmp/data/files/debian-stretch-subutai-template_0.4.1_amd64.tar.gz", "md5")
		sha256Sum := Hash("/tmp/data/files/debian-stretch-subutai-template_0.4.1_amd64.tar.gz", "sha256")
		os.Rename("/tmp/data/files/debian-stretch-subutai-template_0.4.1_amd64.tar.gz", "/tmp/data/files/" + md5Sum)
		tests[5].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "debian-stretch-subutai-template_0.4.1_amd64.tar.gz",
			Repo:     "template",
			Owner:    SubutaiName,
			Token:    SubutaiToken,
			Private:  "false",
			Tags:     "nobodyreadstags,haha",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[5].wantErr = false
	}
	{
		cmd := exec.Command("wget", "-O", "/tmp/data/public/akenzhaliev/generic-ansible-subutai-template_0.4.1_amd64.tar.gz", "https://cdn.subutai.io:8338/kurjun/rest/raw/download?id=cde85454-e6a9-42f6-8447-fc09ad96249b")
		cmd.Run()
		log.Info("Downloaded file")
		actualFile, _ := os.Open("/tmp/data/public/akenzhaliev/generic-ansible-subutai-template_0.4.1_amd64.tar.gz")
		auxFile, _ := os.Create("/tmp/data/files/generic-ansible-subutai-template_0.4.1_amd64.tar.gz")
		io.Copy(auxFile, actualFile)
		auxFileStats, _ := os.Stat("/tmp/data/files/generic-ansible-subutai-template_0.4.1_amd64.tar.gz")
		md5Sum := Hash("/tmp/data/files/generic-ansible-subutai-template_0.4.1_amd64.tar.gz", "md5")
		sha256Sum := Hash("/tmp/data/files/generic-ansible-subutai-template_0.4.1_amd64.tar.gz", "sha256")
		os.Rename("/tmp/data/files/generic-ansible-subutai-template_0.4.1_amd64.tar.gz", "/tmp/data/files/" + md5Sum)
		tests[6].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "generic-ansible-subutai-template_0.4.1_amd64.tar.gz",
			Repo:     "template",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "false",
			Tags:     "nobodyreadstags",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[6].wantErr = false
	}
	{
		cmd := exec.Command("wget", "-O", "/tmp/data/public/akenzhaliev/management-subutai-template_7.0.2_amd64.tar.gz", "https://cdn.subutai.io:8338/kurjun/rest/raw/download?id=14644624-4f09-41a7-8168-5a5b9668adab")
		cmd.Run()
		log.Info("Downloaded file")
		actualFile, _ := os.Open("/tmp/data/public/akenzhaliev/management-subutai-template_7.0.2_amd64.tar.gz")
		auxFile, _ := os.Create("/tmp/data/files/management-subutai-template_7.0.2_amd64.tar.gz")
		io.Copy(auxFile, actualFile)
		auxFileStats, _ := os.Stat("/tmp/data/files/management-subutai-template_7.0.2_amd64.tar.gz")
		md5Sum := Hash("/tmp/data/files/management-subutai-template_7.0.2_amd64.tar.gz", "md5")
		sha256Sum := Hash("/tmp/data/files/management-subutai-template_7.0.2_amd64.tar.gz", "sha256")
		os.Rename("/tmp/data/files/management-subutai-template_7.0.2_amd64.tar.gz", "/tmp/data/files/" + md5Sum)
		tests[7].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "management-subutai-template_7.0.2_amd64.tar.gz",
			Repo:     "template",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "true",
			Tags:     "nobodyreadstags",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[7].wantErr = false
	}
	{
		cmd := exec.Command("wget", "-O", "/tmp/data/public/akenzhaliev/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz", "https://cdn.subutai.io:8338/kurjun/rest/template/download?id=e135003f-fb13-47e7-8485-15f5ba7f1af4")
		cmd.Run()
		log.Info("Downloaded file")
		md5Sum := Hash("/tmp/data/public/akenzhaliev/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz", "md5")
		sha256Sum := Hash("/tmp/data/public/akenzhaliev/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz", "sha256")
		tests[8].fields = fields {
			Filename: "nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz",
			Repo:     "template",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "true",
			Tags:     "nobodyreadstags",
			md5:      md5Sum,
			sha256:   sha256Sum,
		}
		tests[8].wantErr = true
	}
	{
		actualFile, _ := os.Open("/tmp/data/public/akenzhaliev/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz")
		auxFile, _ := os.Create("/tmp/data/files/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz")
		io.Copy(auxFile, actualFile)
		auxFileStats, _ := os.Stat("/tmp/data/files/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz")
		md5Sum := Hash("/tmp/data/public/akenzhaliev/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz", "md5")
		sha256Sum := Hash("/tmp/data/public/akenzhaliev/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz", "sha256")
		os.Rename("/tmp/data/files/nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz", "/tmp/data/files/" + md5Sum)
		tests[9].fields = fields {
			File:     io.Reader(auxFile),
			Filename: "nogeneric-ansible-subutai-template_0.4.1_amd64.tar.gz",
			Repo:     "template",
			Owner:    AkenzhalievName,
			Token:    AkenzhalievToken,
			Private:  "true",
			Tags:     "nobodyreadstags",
			md5:      md5Sum,
			sha256:   sha256Sum,
			size:     auxFileStats.Size(),
		}
		tests[9].wantErr = true
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
			if got := request.BuildResult(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UploadRequest.BuildResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadRequest_HandlePrivate(t *testing.T) {
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
			request.HandlePrivate()
		})
	}
}

func TestUploadRequest_ReadDeb(t *testing.T) {
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
			gotControl, err := request.ReadDeb()
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.ReadDeb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotControl, tt.wantControl) {
				t.Errorf("UploadRequest.ReadDeb() = %v, want %v", gotControl, tt.wantControl)
			}
		})
	}
}

func TestGetControl(t *testing.T) {
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
			if err := request.TemplateCheckValid(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_TemplateCheckFieldsPresent(t *testing.T) {
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
			if err := request.TemplateCheckFormat(tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.TemplateCheckFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadRequest_UploadTemplate(t *testing.T) {
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
			if err := request.Upload(); (err != nil) != tt.wantErr {
				t.Errorf("UploadRequest.Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
