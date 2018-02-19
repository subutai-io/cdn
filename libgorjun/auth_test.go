package gorjun

import (
	"fmt"
	"testing"
	"time"
	"math/rand"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

var r *rand.Rand // Rand for this package.

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func TestGorjunServer_AuthenticateUser(t *testing.T) {
	g := NewGorjunServer();
	_, err := g.RegisterUser("tester", "publickey")
	if err != nil {
		t.Errorf("Failed to register user: %v", err)
	}
}

func TestGorjunServer_RegisterUserWithMultipleKeys(t *testing.T) {
	g := NewGorjunServer();
	randomUserName := RandomString(10)
	for i:= 1; i <= 100; i++ {
		randomKey := RandomString(100)
		_, err := g.RegisterUser(randomUserName, randomKey)
		if err != nil {
			t.Errorf("Failed to register user: %v", err)
		}
		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/auth/keys?user=%s", g.Hostname, randomUserName))
		if err != nil {
			fmt.Errorf("Failed to retrieve user keys: %v", err)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
		}
		var f []GorjunFile
		err = json.Unmarshal(data, &f)
		if err != nil {
			fmt.Errorf("Failed to unmarshal contents from %s: %v", g.Hostname, err)
		}
		assert.Equal(t, i, len(f), "Numbers of key should be equal")
	}
}

func TestGorjunServer_GetKeysByOwner(t *testing.T) {
	g := NewGorjunServer();
	artifactTypes := [3]string{"template", "raw", "apt"}
	for i:= 0; i < len(artifactTypes); i++ {
		resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/" + artifactTypes[i] + "/list", g.Hostname))
		if err != nil {
			fmt.Errorf("Failed to retrieve list of %s  s: %v", artifactTypes[i], err)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		var artifacts []GorjunFile
		err = json.Unmarshal(data, &artifacts)
		for j:= 0; j < len(artifacts); j++ {
			if len(artifacts[j].Owner) > 0 {
				resp, _ := http.Get(fmt.Sprintf("http://%s/kurjun/rest/auth/keys?user=%s", g.Hostname, artifacts[j].Owner[0]))
				data, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				var keys []Keys
				err = json.Unmarshal(data, &keys)
				assert.NotEqual(t, len(keys), 0, "Keys of existing user should be greater than zero")
			}
		}
	}

}
func TestGetAuthTokenCode(t *testing.T) {
	g := NewGorjunServer();
	err := g.GetAuthTokenCode()
	if err != nil {
		t.Errorf("Failed to retrieve token: %v", err)
	}
	if len(g.TokenCode) != 32 {
		t.Errorf("Token length doesn't equals 32 symbols: %d", len(g.TokenCode))
	}
}

func TestGetActiveToken(t *testing.T) {
	g := NewGorjunServer()
	err := g.GetAuthTokenCode()
	if err != nil {
		t.Errorf("Failed to retrieve token: %v", err)
	}
	fmt.Printf("Token code: %s\n", g.TokenCode)
	sign, err := g.SignToken(g.TokenCode)
	if err != nil {
		t.Errorf("Failed to sign token code: %v", err)
	}
	fmt.Printf("Signed token code: %s\n", sign)
	err = g.GetActiveToken(sign)
	if err != nil {
		t.Errorf("Failed to get active token: %v", err)
	}
	fmt.Printf("Active token: %s, len: %d\n", g.Token, len(g.Token))
}
