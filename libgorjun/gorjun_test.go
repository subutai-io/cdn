package gorjun

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestListUserFiles(t *testing.T) {
	g := NewGorjunServer()
	err := g.AuthenticateUser()
	if err != nil {
		t.Errorf("Authnetication failure: %v", err)
	}

	d1 := []byte("This is a test file\n")
	ioutil.WriteFile("/tmp/libgorjun-test", d1, 0644)
	id, err := g.Upload("/tmp/libgorjun-test", "raw", "false")
	if err != nil {
		t.Errorf("Failed to upload: %v", err)
	}
	fmt.Printf("File ID: %s", id)

	flist, err := g.ListUserFiles()
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	if len(flist) <= 0 {
		t.Errorf("Resulting array is empty")
	}
	err = g.Deletes("raw", "")
	if err != nil {
		t.Errorf("Failed to delete raw files: %v", err)
	}
}

func TestRemoveTemplate(t *testing.T) {
	g := NewGorjunServer()
	err := g.AuthenticateUser()
	if err != nil {
		t.Errorf("Authnetication failure: %v", err)
	}
	id, err := g.Upload("data/abdysamat-apache-subutai-template_4.0.0_amd64.tar.gz", "template", "false")
	if err != nil {
		t.Errorf("Failed to upload: %v", err)
	}
	fmt.Printf("Template uploaded successfully, id : %s\n", id)
	err = g.RemoveFileByID(id, "template")
	if err != nil {
		t.Errorf("Failed to remove file: %v", err)
	}
	fmt.Printf("Template removed successfully, id : %s\n", id)
}

//TestGorjunServer_CheckTemplatesSignatureExist will check signatures of
//templates, all templates should have more than zero signatures
func TestGorjunServer_CheckTemplatesSignatureExist(t *testing.T) {
	g := NewGorjunServer()
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/list", g.Hostname))
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	var templates []GorjunFile
	err = json.Unmarshal(data, &templates)
	for _, template := range templates {
		fmt.Printf("ID of templates is %s\n", template.ID)
		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/info?id=%s", g.Hostname, template.ID))
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
		}
		var templateInfo []GorjunFile
		err = json.Unmarshal(data, &templateInfo)
		fmt.Printf("Len of signatture  is %d\n", len(templateInfo[0].Signature))
		assert.NotEqual(t, len(templateInfo[0].Signature), 0, "Template with ID = %s should have signature\n", template.ID)
	}
}

func Shuffle(a []string) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

//TestGorjunServer_GetLatestTemplateByVersion will upload templates
//with different version in random order, info rest should return latest by version
//if several version exits it should return by date
func TestGorjunServer_GetLatestTemplateByVersion(t *testing.T) {
	g := NewGorjunServer()
	err := g.AuthenticateUser()
	if err != nil {
		t.Errorf("Authnetication failure: %v", err)
	}
	var dates []int
	templateVersions := []string{"0.1.6", "0.1.7", "0.1.9", "0.1.10", "0.1.11"}
	rand.Seed(time.Now().UnixNano())
	Shuffle(templateVersions)

	for _, version := range templateVersions {
		id, err := g.Upload("data/nginx-subutai-template_"+version+"_amd64.tar.gz", "template", "false")
		if err != nil {
			t.Errorf("Failed to upload: %v", err)
		}
		fmt.Printf("Template uploaded successfully, id : %s\n", id)

		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/info?id=%s", g.Hostname, id))
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
		}
		var template []GorjunFile
		err = json.Unmarshal(data, &template)
		timestamp, err := strconv.Atoi(template[0].Timestamp)
		dates = append(dates, timestamp)
		time.Sleep(100 * time.Millisecond)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/info?name=%s&owner=%s", g.Hostname, "nginx", g.Username))
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	var template []GorjunFile
	err = json.Unmarshal(data, &template)
	assert.Equal(t, "0.1.11", template[0].Version)
	g.Deletes("template", "")
}

//TestGorjunServer_GetLatestRaw will upload raw
//files , info rest should return  by date
func TestGorjunServer_GetLatestRaw(t *testing.T) {
	g := NewGorjunServer()
	err := g.AuthenticateUser()
	if err != nil {
		t.Errorf("Authnetication failure: %v", err)
	}
	var dates []int
	rawNumber := 10

	for i := 1; i <= rawNumber; i++ {
		id, err := g.Upload("data/abdysamat-apache-subutai-template_4.0.0_amd64.tar.gz", "raw", "false")
		if err != nil {
			t.Errorf("Failed to upload: %v", err)
		}
		fmt.Printf("Raw uploaded successfully, id : %s\n", id)

		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/raw/info?id=%s", g.Hostname, id))
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
		}
		var template []GorjunFile
		err = json.Unmarshal(data, &template)
		timestamp, err := strconv.Atoi(template[0].Timestamp)
		dates = append(dates, timestamp)
		time.Sleep(101 * time.Millisecond)
	}
	sort.Ints(dates)
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/raw/info?name=%s&owner=%s", g.Hostname, "abdysamat-apache", g.Username))
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	var template []GorjunFile
	err = json.Unmarshal(data, &template)
	timestamp, err := strconv.Atoi(template[0].Timestamp)
	fmt.Println(dates)
	fmt.Println(timestamp)
	fmt.Println(dates[rawNumber-1])
	assert.Equal(t, timestamp, dates[rawNumber-1])
	g.Deletes("raw", "")
}

//TestGorjunServer_SameTemplateUpload will upload
//same template twice, old template should deleted
func TestGorjunServer_SameTemplateUpload(t *testing.T) {
	g := NewGorjunServer()
	templateVersions := []string{"0.1.6", "0.1.7", "0.1.9", "0.1.10", "0.1.11"}
	for _, version := range templateVersions {
		resp, _ := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/info?name=%s&version=%s", g.Hostname, "nginx", version))
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Test can't be run because templates should uploaded")
			return
		}
	}
	TestGorjunServer_GetLatestTemplateByVersion(t)
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/list", g.Hostname))
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	var templateList []GorjunFile
	err = json.Unmarshal(data, &templateList)

	m := make(map[string]int)

	for _, template := range templateList {
		s := template.Owner[0] + template.Name + template.Version
		m[s]++
		assert.NotEqual(t, m[s], 2, "Same template exist twice", template.ID)
	}
}
