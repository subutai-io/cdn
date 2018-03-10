package gorjun

import (
	//"encoding/json"
	//"fmt"
	//"io/ioutil"
	//"math/rand"
	//"net/http"
	//"os/exec"
	//"strconv"
	//"testing"
	//"time"
	//"github.com/stretchr/testify/assert"
)

//TestGorjunServer_CheckingFilesAfterDeleting will upload templates
//after will delete all templates and outputs files on gorjun directory
//TODO same problem
//func TestGorjunServer_CheckingFilesAfterDeleting(t *testing.T) {
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
//		id, err := g.Upload("data/nginx-subutai-template_"+version+"_amd64.tar.gz", "template","false")
//		if err != nil {
//			t.Errorf("Failed to upload: %v", err)
//		}
//		fmt.Printf("Template uploaded successfully, id : %s\n", id)
//
//		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/info?id=%s", g.Hostname, id))
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
//	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/info?name=%s&owner=%s", g.Hostname, "nginx", g.Username))
//	data, err := ioutil.ReadAll(resp.Body)
//	resp.Body.Close()
//	if err != nil {
//		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
//	}
//
//	resp, err = http.Get(fmt.Sprintf("http://%s/kurjun/rest/template/list", g.Hostname))
//	data, err = ioutil.ReadAll(resp.Body)
//	resp.Body.Close()
//	if err != nil {
//		fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
//	}
//	var templateList []GorjunFile
//	err = json.Unmarshal(data, &templateList)
//
//	for _, template := range templateList {
//		err = g.RemoveFileByID(template.ID, "template")
//		if err != nil {
//			t.Errorf("Failed to remove file: %v", err)
//		}
//		fmt.Printf("Template removed successfully, id : %s\n", template.ID)
//	}
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting templates \n%s\n", output)
//	assert.Equal(t, 0, len(output))
//}


//TestGorjunServer_TwoUserUploadsSameTemplate two different
//user will upload same template
//First user will delete his template, second user won't
//Second user should able to download his template
//TODO This test not solved yet, also
//func TestGorjunServer_TwoUserUploadsSameTemplate(t *testing.T) {
//	g := NewGorjunServer()
//	output, _ := exec.Command("bash", "-c", "gpg --armor --export tester").Output()
//	g.RegisterUser(g.Username, string(output))
//	err := g.AuthenticateUser()
//	if err != nil {
//		t.Errorf("Authnetication failure: %v", err)
//	}
//
//	idFirstTemplate, err := g.Upload("data/abdysamat-apache-subutai-template_4.0.0_amd64.tar.gz", "template","false")
//	if err != nil {
//		t.Errorf("Failed to upload: %v", err)
//	}
//	fmt.Printf("Template uploaded successfully, id : %s\n", idFirstTemplate)
//
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting templates \n%s\n", output)
//	assert.NotEqual(t, 0, len(output))
//
//	g = NewGorjunServer()
//	output, _ = exec.Command("bash", "-c", "gpg --armor --export emilbeksulaymanov").Output()
//	g.Username = "emilbeksulaymanov"
//	g.RegisterUser(g.Username, string(output))
//	err = g.AuthenticateUser()
//	if err != nil {
//		t.Errorf("Authnetication failure: %v", err)
//	}
//
//	idSecondTemplate, err := g.Upload("data/abdysamat-apache-subutai-template_4.0.0_amd64.tar.gz", "template","false")
//	if err != nil {
//		t.Errorf("Failed to upload: %v", err)
//	}
//	fmt.Printf("Template uploaded successfully, id : %s\n", idSecondTemplate)
//
//	//err = g.RemoveFileByID(idFirstTemplate, "template")
//	//if err != nil {
//	//	t.Errorf("Failed to remove file: %v", err)
//	//}
//	//fmt.Printf("Template removed successfully, id : %s\n", idFirstTemplate)
//
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting templates \n%s\n", output)
//	assert.NotEqual(t, 0, len(output))
//}

//TestGorjunServer_UploadOneFilesToAllRepos will upload
//one deb file to 2 repo, and will delete one by one
//TODO This test not solved yet
//func TestGorjunServer_UploadOneFilesToAllRepos(t *testing.T) {
//	g := NewGorjunServer()
//	output, _ := exec.Command("bash", "-c", "gpg --armor --export tester").Output()
//	g.RegisterUser(g.Username, string(output))
//	err := g.AuthenticateUser()
//	if err != nil {
//		t.Errorf("Authnetication failure: %v", err)
//	}
//
//	idRaw, err := g.Upload("data/winff_1.5.5-1_all.deb", "raw","false")
//	if err != nil {
//		t.Errorf("Failed to upload: %v", err)
//	}
//	fmt.Printf("Raw uploaded successfully, id : %s\n", idRaw)
//
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting raw files \n%s\n", output)
//	assert.NotEqual(t, 0, len(output))
//
//	idDeb, err := g.Upload("data/winff_1.5.5-1_all.deb", "apt","false")
//	if err != nil {
//		t.Errorf("Failed to upload: %v", err)
//	}
//	fmt.Printf("Apt uploaded successfully, id : %s\n", idDeb)
//
//	err = g.RemoveFileByID(idRaw, "raw")
//	if err != nil {
//		t.Errorf("Failed to remove file: %v", err)
//	}
//	fmt.Printf("Raw removed successfully, id : %s\n", idRaw)
//
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting raw files \n%s\n", output)
//	assert.NotEqual(t, 0, len(output))
//
//	err = g.RemoveFileByID(idDeb, "apt")
//	if err != nil {
//		t.Errorf("Failed to remove file: %v", err)
//	}
//	fmt.Printf("Apt removed successfully, id : %s\n", idDeb)
//
//	output, _ = exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
//	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting deb files \n%s\n", output)
//	assert.Equal(t, 0, len(output))
//}