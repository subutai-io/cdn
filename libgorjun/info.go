package gorjun

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"log"
	"os/exec"
	"time"
	"math/rand"
	"bytes"
)

// List returns a list of files
func (g *GorjunServer) List(artifactType string, parameters string) ([]GorjunFile, error) {
	resp, err:= http.Get(fmt.Sprintf("http://%s/kurjun/rest/" + artifactType + "/info", g.Hostname))
	if len(parameters) != 0 {
		resp, err = http.Get(fmt.Sprintf("http://%s/kurjun/rest/" + artifactType + "/info" + parameters, g.Hostname))
	}
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
		log.Printf("error decoding response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("response: %q", data)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal contents from %s: %v", g.Hostname, err)
	}
	return rf, nil
}

//Upload files
func (g *GorjunServer) Uploads(artifacttype, private string) (error){
	err := g.AuthenticateUser()
	if err != nil {
		fmt.Errorf("Authnetication failure: %v", err)
	}

	templateVersions := []string{"0.1.6", "0.1.7", "0.1.9", "0.1.10", "0.1.11"}
	rand.Seed(time.Now().UnixNano())

	for _, version := range templateVersions {
		id, err := g.Upload("data/nginx-subutai-template_"+version+"_amd64.tar.gz", artifacttype, private)
		if err != nil {
			fmt.Errorf("Failed to upload: %v", err)
		}
		fmt.Printf("%s uploaded successfully, id : %s\n",artifacttype, id)
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

//Delete files
func (g *GorjunServer) Deletes(artifacttype,parameters string) (error){
	err := g.AuthenticateUser()
	if err != nil {
		fmt.Errorf("Authnetication failure: %v", err)
	}

	flist, err := g.List(artifacttype,parameters)
	for _, template := range flist {
		err = g.RemoveFileByID(template.ID, artifacttype)
		if err != nil {
			fmt.Errorf("Failed to remove file: %v", err)
		}
		fmt.Printf("%s removed successfully, id : %s\n", artifacttype,template.ID)
	}
	showFileSystemState()
	return err
}

func showFileSystemState()  {
	output, _ := exec.Command("bash", "-c", " ls /opt/gorjun/data/files/").Output()
	fmt.Printf("\nList of files in /opt/gorjun/data/files/ directory after deleting templates \n%s\n", output)
}

//Share files
func (g *GorjunServer) Share(token string, lists[]GorjunFile, shareWith, artifactType string) (error){
	err := g.AuthenticateUser()
	if err != nil {
		fmt.Errorf("Authnetication failure: %v", err)
	}

	type share struct {
		Token  string   `json:"token"`
		Id     string   `json:"id"`
		Add    []string `json:"add"`
		Remove []string `json:"remove"`
		Repo   string   `json:"repo"`
	}

	for _, list := range lists {
		locJson, err := json.Marshal(share{Token: token, Id: list.ID, Add: []string{shareWith}, Remove: []string{}, Repo: artifactType})

		req, err := http.NewRequest("POST", "http://127.0.0.1:8080/kurjun/rest/share", bytes.NewBuffer(locJson))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)

		fmt.Println("Response: ", string(body))
		resp.Body.Close()
	}
	return err
}