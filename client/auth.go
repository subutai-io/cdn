package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"mime/multipart"
	"os/exec"
	"github.com/subutai-io/agent/log"
)

func (g *GorjunUser) RegisterUser(username string, publicKey string) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormField("name")
	if err != nil {
		return "", fmt.Errorf("Failed to create name form: %v", err)
	}
	if _, err = fw.Write([]byte(username)); err != nil {
		return "", fmt.Errorf("Failed to write token: %v", err)
	}
	if fw, err = w.CreateFormField("key"); err != nil {
		return "", fmt.Errorf("Failed to create key form field: %v", err)
	}
	if _, err = fw.Write([]byte(publicKey)); err != nil {
		return "", fmt.Errorf("Failed to write token: %v", err)
	}
	w.Close()
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/kurjun/rest/auth/register", g.Hostname), &b)
	if err != nil {
		return "", fmt.Errorf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to execute HTTP request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Registration failed. Server returned %s error", res.Status)
	}
	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}
	return string(response), nil
}

func (g *GorjunUser) Register(username string)  {
	output, _ := exec.Command("bash", "-c", "gpg --armor --export " + username).Output()
	g.RegisterUser(g.Username, string(output))
}

// AuthenticateUser will try to authenticate user by downloading his token code, signing it with GPG
// and sending it back to server to get user token
// If passphrase is not empty, PGP will try to decrypt the private key before signing the code
// if gpgdir is empty, the default ($HOME/.gnupg) will be used
func (g *GorjunUser) AuthenticateUser() error {
	err := g.GetAuthTokenCode()
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't get AuthID for %s", g.Username))
		return err
	}
	log.Info(fmt.Sprintf("Signing AuthID: %s", g.TokenCode))
	sign, err := g.SignToken(g.TokenCode)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't sign AuthID for %s", g.Username))
		return err
	}
	err = g.GetActiveToken(sign)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't get token for %s", g.Username))
		return err
	}
	return nil
}

// GetAuthTokenCode is a first step of authentication - it requests a special code from the server.
// This code needs to be PGP-signed later
func (g *GorjunUser) GetAuthTokenCode() error {
	fmt.Println("Getting auth id for user " + g.Username)
	resp, err := http.Get(fmt.Sprintf("http://%s/kurjun/rest/auth/token?user=%s", g.Hostname, g.Username))
	if err != nil {
		return fmt.Errorf("Failed to retrieve unsigned token: %v", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	g.TokenCode = string(data)
	return nil
}

// GetActiveToken will send signed message to server and return active token
// that will be used for authneticated requests
func (g *GorjunUser) GetActiveToken(signed string) error {
	signed = "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA256\n\n" + g.TokenCode + "\n" + signed + "\n"
	form := url.Values{
		"message": {signed},
		"user":    {g.Username},
	}
	body := bytes.NewBufferString(form.Encode())
	resp, err := http.Post(fmt.Sprintf("http://%s/kurjun/rest/auth/token", g.Hostname), "application/x-www-form-urlencoded", body)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to retrieve active token: %v", err))
		return fmt.Errorf("Failed to retrieve active token: %v", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to read body from %s: %v", g.Hostname, err))
		return fmt.Errorf("Failed to read body from %s: %v", g.Hostname, err)
	}
	g.Token = string(data)
	if len(g.Token) != 64 {
		reason := g.Token
		g.Token = ""
		log.Warn(fmt.Sprintf("Failed to retrieve active token: %s", reason))
		return fmt.Errorf("Failed to retrieve active token: %s", reason)
	}
	return nil
}

func (g *GorjunUser) GetKeyByEmail(keyring openpgp.EntityList, email string) *openpgp.Entity {
	for _, entity := range keyring {
		for _, ident := range entity.Identities {
			if ident.UserId.Email == email {
				return entity
			}
		}
	}
	return nil
}

// SignToken will sign with GnuPG provided token and return signed version
func (g *GorjunUser) SignToken(token string) (string, error) {
	if g.GPGDirectory == "" {
		return "", fmt.Errorf("GPG Directory was not specified")
	}
	// GPG may have two variants of key storage - in secring.gpg/pubring.gpg for older versions
	// and for pubring.kbx and separate directory for private key in version of GnuPG 2.1+
	pubringPath := g.GPGDirectory + "/pubring.gpg"
	if _, err := os.Stat(pubringPath); os.IsNotExist(err) {
		pubringPath = g.GPGDirectory + "/pubring.kbx"
	}
	if _, err := os.Stat(pubringPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Can't find pubring.gpg nor pubring.kbx")
	}
	pukFile, err := os.Open(g.GPGDirectory + "/pubring.gpg")
	defer pukFile.Close()
	if err != nil {
		return "", fmt.Errorf("Failed to open public keyring file: %v", err)
	}
	pubring, err := openpgp.ReadKeyRing(pukFile)
	if err != nil {
		return "", fmt.Errorf("Failed to read public keyring: %v", err)
	}
	publicKey := g.GetKeyByEmail(pubring, g.Email)
	if publicKey == nil {
		return "", fmt.Errorf("Public key for %s was not found", g.Email)
	}

	priFile, err := os.Open(g.GPGDirectory + "/secring.gpg")
	defer priFile.Close()
	if err != nil {
		return "", fmt.Errorf("Failed to open private keyring file: %v", err)
	}
	secring, err := openpgp.ReadKeyRing(priFile)
	if err != nil {
		return "", fmt.Errorf("Failed to read private keyring: %v", err)
	}
	privateKey := g.GetKeyByEmail(secring, g.Email)
	if privateKey == nil {
		return "", fmt.Errorf("Private key for %s was not found", g.Email)
	}
	if g.Passphrase != "" {
		privateKey.PrivateKey.Decrypt([]byte(g.Passphrase))
	}
	outBuf := new(bytes.Buffer)
	err = openpgp.ArmoredDetachSign(outBuf, privateKey, strings.NewReader(token), nil)
	if err != nil {
		return "", fmt.Errorf("Failed to sign token: %s", err)
	}
	return outBuf.String(), nil
}

func (g *GorjunUser) decodePrivateKey() (*packet.PrivateKey, error) {
	in, err := os.Open(g.GPGDirectory + "/secring.gpg")
	if err != nil {
		in.Close()
		return nil, err
	}
	block, err := armor.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode GPG Armor: %s", err)
	}
	if block.Type != openpgp.PrivateKeyType {
		return nil, fmt.Errorf("Invalid private key file")
	}
	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		return nil, fmt.Errorf("Error reading private key")
	}
	key, success := pkt.(*packet.PrivateKey)
	if !success {
		return nil, fmt.Errorf("Error parsing private key")
	}
	return key, nil
}