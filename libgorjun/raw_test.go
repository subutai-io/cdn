package gorjun

import (
	//"encoding/json"
	//"strconv"
	//"time"
	//"fmt"
	//"io/ioutil"
	//"os/exec"
	//"github.com/stretchr/testify/assert"
	//"testing"
	//"math/rand"
	//"net/http"
)

//TestGorjunServer_CheckingRawFilesAfterDeleting will upload raw foles
//after will delete all raw files and outputs files on gorjun directory
//TODO raw
//func TestGorjunServer_CheckingRawFilesAfterDeleting(t *testing.T) {
//	g := NewGorjunServer()
//	output, _ := exec.Command("bash", "-c", "gpg --armor --export tester").Output()
//	g.RegisterUser(g.Username, string(output))
//	err := g.AuthenticateUser()
//	if err != nil {
//		t.Errorf("Authnetication failure: %v", err)
//	}
//	var dates []int
//	templateVersions := []string{"0.1.6", "0.1.7", "0.1.9", "0.1.10", "0.1.11"}
//	rand.Seed(time.Now().UnixNano())
//	Shuffle(templateVersions)
//
//	for _, version := range templateVersions {
//		id, err := g.Upload("data/nginx-subutai-template_"+version+"_amd64.tar.gz", "raw","false")
//		if err != nil {
//			t.Errorf("Failed to upload: %v", err)
//		}
//		fmt.Printf("Raw uploaded successfully, id : %s\n", id)
//
//		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/raw/info?id=%s", g.Hostname, id))
//		data, err := ioutil.ReadAll(resp.Body)
//		resp.Body.Close()
//		if err != nil {
//			fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
//		}
//		var template []GorjunFile
//		err = json.Unmarshal(data, &template)
//		timestamp, err := strconv.Atoi(template[0].Timestamp)
//		dates = append(dates, timestamp)
//		time.Sleep(100 * time.Millisecond)
//	}
//	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/raw/info?name=%s&owner=%s", g.Hostname, "nginx", g.Username))
//	data, err := ioutil.ReadAll(resp.Body)
//	resp.Body.Close()
//	if err != nil {
//		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
//	}
//
//	resp, err = http.Get(fmt.Sprintf("http://%s/kurjun/rest/raw/list", g.Hostname))
//	data, err = ioutil.ReadAll(resp.Body)
//	resp.Body.Close()
//	if err != nil {
//		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
//	}
//	var templateList []GorjunFile
//	err = json.Unmarshal(data, &templateList)
//
//	for _, template := range templateList {
//		err = g.RemoveFileByID(template.ID, "raw")
//		if err != nil {
//			t.Errorf("Failed to remove file: %v", err)
//		}
//		fmt.Printf("Raw removed successfully, id : %s\n", template.ID)
//	}
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting raw files \n%s\n", output)
//	assert.Equal(t, 0, len(output))
//}