package gorjun

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test for info rest
// Two user uploads same 5 templates
// Each user requests by his token
func TestList(t *testing.T) {
	g := NewGorjunServer()
	g.Register(g.Username)

	k := SecondNewGorjunServer()
	k.Register(k.Username)

	v := VerifiedUser()
	v.Register(v.Username)

	repos := [2]string{"template", "raw"}

	for i := 0; i < len(repos); i++ {

		artifactType := repos[i]

		err := k.Uploads(artifactType, "false")
		if err != nil {
			t.Errorf("Failed to uploads templates: %v", err)
		}

		err = g.Uploads(artifactType, "false")

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

		err = v.Deletes(artifactType, "")
		if err != nil {
			t.Errorf("Failed to delete templates: %v", err)
		}
	}
}

// Test for info rest
// Two user uploads same 5 private templates
// Each user requests by his token
func TestPrivateTemplates(t *testing.T) {
	g := NewGorjunServer()
	g.Register(g.Username)

	k := SecondNewGorjunServer()
	k.Register(k.Username)

	v := VerifiedUser()
	v.Register(g.Username)

	repos := [2]string{"template", "raw"}

	for i := 0; i < len(repos); i++ {

		artifactType := repos[i]

		err := k.Uploads(artifactType, "true")
		if err != nil {
			t.Errorf("Failed to uploads templates: %v", err)
		}

		err = g.Uploads(artifactType, "true")
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

		err = v.Deletes(artifactType, "?token="+g.Token)
		if err != nil {
			t.Errorf("Failed to delete templates: %v", err)
		}
	}
}

// Test for info rest
// If can't find user templates it should search
// in shared
func TestShareTemplates(t *testing.T) {
	g := NewGorjunServer()
	g.Register(g.Username)

	k := SecondNewGorjunServer()
	k.Register(k.Username)
	k.AuthenticateUser()

	artifactType := "template"

	err := g.Uploads(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads templates: %v", err)
	}

	fmt.Println("Token for user " + g.Username + " = " + g.Token)
	fmt.Println("Token for user " + k.Username + " = " + k.Token)
	flist, err := g.List(artifactType, "")
	if err != nil {
		t.Errorf("Failed to retrieve user files: %v", err)
	}
	g.Share(g.Token, flist, k.Username, artifactType)

	err = g.Deletes(artifactType, "?token="+g.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

//Test for info?name=master
func TestListByName(t *testing.T) {
	v := VerifiedUser()
	v.Register(v.Username)

	g := SecondNewGorjunServer()
	g.Register(g.Username)

	artifactType := "template"

	err := g.Uploads(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}

	err = v.Uploads(artifactType, "false")
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
	err = v.Deletes(artifactType, "?token="+v.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

//Test for info?name=master&version=
func TestListByVersion(t *testing.T) {
	v := VerifiedUser()
	v.Register(v.Username)

	g := SecondNewGorjunServer()
	g.Register(g.Username)

	artifactType := "template"

	err := g.Uploads(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}

	err = v.Uploads(artifactType, "false")
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

	err = v.Deletes(artifactType, "?token="+v.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

/*
user1 and user2
each have private template my-template
1) use name+token, each user must find own template
2) use name+owner+token, specify another user as owner, each user must find another user template
each user have public template my-pub-template
1) user name+token, each user must find latest added my-pub-template
2) use name+owner+token, specify another user as owner, each user must find another user template
remove user1 private and public templates
1) use name+token, both users find user2 public template
only user2 finds user2 private template
2) use name+owner+token,  specify another user as owner, only user2 templates are found
vse
*/
func TestListHardTest(t *testing.T) {
	k := NewGorjunServer()
	k.Register(k.Username)

	g := SecondNewGorjunServer()
	g.Register(g.Username)

	v := VerifiedUser()
	v.Register(v.Username)

	artifactType := "template"

	err := g.Uploads(artifactType, "false")
	if err != nil {
		t.Errorf("Failed to uploads %s: %v", err, artifactType)
	}

	err = k.Uploads(artifactType, "false")
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

	err = k.Deletes(artifactType, "?token="+v.Token)
	if err != nil {
		t.Errorf("Failed to delete templates: %v", err)
	}
}

//Search should work in this order
func TestListPriority(t *testing.T) {
	v := VerifiedUser()
	v.Register(v.Username)
	v.AuthenticateUser()
	g := NewGorjunServer()
	g.Register(g.Username)

	//artifactType := "template"
	v.Upload("data/ceph-subutai-template_4.0.0_amd64.tar.gz", "template", "false")
	v.Upload("data/master-subutai-template_4.0.0_amd64.tar.gz", "template", "false")

	//err := g.Uploads(artifactType, "false")
	//if err != nil {
	//	t.Errorf("Failed to uploads %s: %v", err, artifactType)
	//}
	fmt.Println("Token for user " + g.Username + " = " + g.Token)
	fmt.Println("Token for user " + v.Username + " = " + v.Token)

}
