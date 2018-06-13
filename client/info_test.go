package client

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
)

// Test for info rest
// Two user uploads same 5 templates
// Each user requests by his token

func TestList(t *testing.T) {
	g := FirstGorjunUser()
	g.Register(g.Username)
	k := SecondGorjunUser()
	k.Register(k.Username)
	v := VerifiedGorjunUser()
	v.Register(v.Username)
	repos := [2]string{"template", "raw"}
	for i := 0; i < len(repos); i++ {
		artifactType := repos[i]
		err := k.TestUpload(artifactType, "false")
		if err != nil {
			t.Errorf("Failed to uploads templates: %v", err)
		}
		err = g.TestUpload(artifactType, "false")
		if err != nil {
			t.Errorf("Failed to uploads templates: %v", err)
		}
		flist, err := g.List(artifactType, "")
		if err != nil {
			t.Errorf("Failed to retrieve user files: %v", err)
		}
		if len(flist) <= 0 {
			t.Errorf("Resulting array is empty")
		}
		flist, err = g.List(artifactType, "?token")
		if err != nil {
			t.Errorf("Failed to retrieve user files: %v", err)
		}
		if len(flist) <= 0 {
			t.Errorf("Resulting array is empty")
		}
		fmt.Println("Token for user " + g.Username + " = " + g.Token)
		fmt.Println("Token for user " + k.Username + " = " + k.Token)
		flist, err = g.List(artifactType, "?name=nginx&owner="+g.Username)
		owner := flist[0].Owner[0]
		assert.Equal(t, owner, g.Username)
		flist, err = g.List(artifactType, "?name=nginx&owner="+k.Username)
		owner = flist[0].Owner[0]
		assert.Equal(t, owner, k.Username)
		flist, err = g.List(artifactType, "?name=nginx&token="+g.Token)
		owner = flist[0].Owner[0]
		assert.Equal(t, owner, g.Username)
		flist, err = g.List(artifactType, "?name=nginx&token="+k.Token)
		owner = flist[0].Owner[0]
		assert.Equal(t, owner, k.Username)
		err = v.Delete(artifactType, "")
		if err != nil {
			t.Errorf("Failed to delete templates: %v", err)
		}
	}
}

// Test for info rest
// Two user uploads same 5 private templates
// Each user requests by his token
func TestPrivateTemplates(t *testing.T) {
	g := FirstGorjunUser()
	g.Register(g.Username)
	k := SecondGorjunUser()
	k.Register(k.Username)
	v := VerifiedGorjunUser()
	v.Register(g.Username)
	repos := [2]string{"template", "raw"}
	for i := 0; i < len(repos); i++ {
		artifactType := repos[i]
		err := k.TestUpload(artifactType, "true")
		if err != nil {
			t.Errorf("Failed to uploads templates: %v", err)
		}
		err = g.TestUpload(artifactType, "true")
		if err != nil {
			t.Errorf("Failed to uploads templates: %v", err)
		}
		fmt.Println("Token for user " + g.Username + " = " + g.Token)
		fmt.Println("Token for user " + k.Username + " = " + k.Token)
		flist, err := g.List(artifactType, "?name=nginx&token="+g.Token)
		owner := flist[0].Owner[0]
		assert.Equal(t, owner, g.Username)
		flist, err = g.List(artifactType, "?name=nginx&token="+k.Token)
		owner = flist[0].Owner[0]
		assert.Equal(t, owner, k.Username)
		err = v.Delete(artifactType, "?token="+g.Token)
		if err != nil {
			t.Errorf("Failed to delete templates: %v", err)
		}
	}
}

// Test for info rest
// If can't find user templates it should search
// in shared
func TestShareTemplates(t *testing.T) {
	g := FirstGorjunUser()
	g.Register(g.Username)
	k := SecondGorjunUser()
	k.Register(k.Username)
	k.AuthenticateUser()
	artifactType := "template"
	err := g.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads templates: %v", err)
	}
	fmt.Println("Token for user " + g.Username + " = " + g.Token)
	fmt.Println("Token for user " + k.Username + " = " + k.Token)
	flist, err := g.List(artifactType, "")
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	g.Share(flist, k.Username, artifactType)
	err = g.Delete(artifactType, "?token="+g.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

// Test for info?name=master
func TestListByName(t *testing.T) {
	v := VerifiedGorjunUser()
	v.Register(v.Username)
	g := SecondGorjunUser()
	g.Register(g.Username)
	artifactType := "template"
	err := g.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}
	err = v.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}
	fmt.Println("Token for user " + g.Username + " = " + g.Token)
	fmt.Println("Token for user " + v.Username + " = " + v.Token)
	flist, err := g.List(artifactType, "?name=nginx")
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner := flist[0].Owner[0]
	version := flist[0].Version
	assert.Equal(t, owner, v.Username)
	assert.Equal(t, version, "0.1.11")
	flist, err = g.List(artifactType, "?name=nginx&token="+g.Token)
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner = flist[0].Owner[0]
	version = flist[0].Version
	assert.Equal(t, owner, g.Username)
	assert.Equal(t, version, "0.1.11")
	flist, err = g.List(artifactType, "?name=nginx&owner="+g.Username)
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner = flist[0].Owner[0]
	version = flist[0].Version
	assert.Equal(t, owner, g.Username)
	assert.Equal(t, version, "0.1.11")
	flist, err = g.List(artifactType, "?name=nginx&owner="+v.Username)
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner = flist[0].Owner[0]
	version = flist[0].Version
	assert.Equal(t, owner, v.Username)
	assert.Equal(t, version, "0.1.11")
	err = v.Delete(artifactType, "?token="+v.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

// Test for info?name=master&version=
func TestListByVersion(t *testing.T) {
	v := VerifiedGorjunUser()
	v.Register(v.Username)
	g := SecondGorjunUser()
	g.Register(g.Username)
	artifactType := "template"
	err := g.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}
	err = v.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}
	fmt.Println("Token for user " + g.Username + " = " + g.Token)
	fmt.Println("Token for user " + v.Username + " = " + v.Token)
	flist, err := g.List(artifactType, "?name=nginx&token="+g.Token+"&version=0.1.6")
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner := flist[0].Owner[0]
	version := flist[0].Version
	assert.Equal(t, owner, g.Username)
	assert.Equal(t, version, "0.1.6")
	flist, err = g.List(artifactType, "?name=nginx&version=0.1.6")
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner = flist[0].Owner[0]
	version = flist[0].Version
	assert.Equal(t, owner, v.Username)
	assert.Equal(t, version, "0.1.6")

	err = v.Delete(artifactType, "?token="+v.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

// user1 and user2
// each have private template my-template
// 1) use name+token, each user must find own template
// 2) use name+owner+token, specify another user as owner, each user must find another user template
// each user have public template my-pub-template
// 1) user name+token, each user must find latest added my-pub-template
// 2) use name+owner+token, specify another user as owner, each user must find another user template
// remove user1 private and public templates
// 1) use name+token, both users find user2 public template
// only user2 finds user2 private template
// 2) use name+owner+token,  specify another user as owner, only user2 templates are found
// vse
func TestListHardTest(t *testing.T) {
	k := FirstGorjunUser()
	k.Register(k.Username)
	g := SecondGorjunUser()
	g.Register(g.Username)
	v := VerifiedGorjunUser()
	v.Register(v.Username)
	artifactType := "template"
	err := g.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}
	err = k.TestUpload(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}
	fmt.Println("Token for user " + g.Username + " = " + g.Token)
	fmt.Println("Token for user " + k.Username + " = " + k.Token)
	flist, err := g.List(artifactType, "?name=nginx")
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	owner := flist[0].Owner[0]
	version := flist[0].Version
	assert.Equal(t, owner, k.Username)
	assert.Equal(t, version, "0.1.11")
	err = k.Delete(artifactType, "?token="+v.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

// Search should work in this order
func TestListPriority(t *testing.T) {
	v := VerifiedGorjunUser()
	v.Register(v.Username)
	v.AuthenticateUser()
	fmt.Printf("Token for user %+v = %+v\n", v.Username, v.Token)
	str, err := v.Upload("data/subutai/debian-stretch-subutai-template_0.4.1_amd64.tar.gz", "template", "false")
	if err != nil {
		fmt.Printf("Could not upload data/subutai/debian-stretch-subutai-template_0.4.1_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	str, err = v.Upload("data/subutai/mysql-subutai-template_0.3.9_amd64.tar.gz", "template", "true")
	if err != nil {
		fmt.Printf("Could not upload data/subutai/mysql-subutai-template_0.3.9_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	str, err = v.Upload("data/subutai/mysql-subutai-template_0.4.0_amd64.tar.gz", "raw", "false")
	if err != nil {
		fmt.Printf("Could not upload data/subutai/mysql-subutai-template_0.4.0_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	str, err = v.Upload("data/subutai/mysql-subutai-template_0.4.1_amd64.tar.gz", "raw", "true")
	if err != nil {
		fmt.Printf("Could not upload data/subutai/mysql-subutai-template_0.4.1_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	f := FirstGorjunUser()
	f.Register(f.Username)
	f.AuthenticateUser()
	fmt.Printf("Token for user %+v = %+v\n", f.Username, f.Token)
	str, err = f.Upload("data/lorem/debian-stretch-subutai-template_0.4.1_amd64.tar.gz", "template", "true")
	if err != nil {
		fmt.Printf("Could not upload data/lorem/debian-stretch-subutai-template_0.4.1_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	str, err = f.Upload("data/lorem/mysql-subutai-template_0.3.9_amd64.tar.gz", "template", "false")
	if err != nil {
		fmt.Printf("Could not upload data/lorem/mysql-subutai-template_0.3.9_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	str, err = f.Upload("data/lorem/mysql-subutai-template_0.4.0_amd64.tar.gz", "raw", "true")
	if err != nil {
		fmt.Printf("Could not upload data/lorem/mysql-subutai-template_0.4.0_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	str, err = f.Upload("data/lorem/mysql-subutai-template_0.4.1_amd64.tar.gz", "raw", "false")
	if err != nil {
		fmt.Printf("Could not upload data/lorem/mysql-subutai-template_0.4.1_amd64.tar.gz: %+v, %+v\n", str, err)
	}
	{
		fileList, _ := v.GetFileByName("mysql-subutai-template_0.3.9_amd64.tar.gz", "template")
		v.Share(fileList, f.Username, "template")
	}
	{
		fileList, _ := f.GetFileByName("mysql-subutai-template_0.4.0_amd64.tar.gz", "raw")
		f.Share(fileList, v.Username, "raw")
	}
}
