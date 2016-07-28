package pgp

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/subutai-io/gorjun/db"

	"github.com/subutai-io/base/agent/log"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

func Verify(name, message string) string {
	entity, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(db.UserKey(name)))
	log.Check(log.WarnLevel, "Reading user public key", err)

	if block, _ := clearsign.Decode([]byte(message)); block != nil {
		_, err = openpgp.CheckDetachedSignature(entity, bytes.NewBuffer(block.Bytes), block.ArmoredSignature.Body)
		if log.Check(log.WarnLevel, "Checking signature", err) {
			return ""
		}
		return string(block.Bytes)
	}
	return ""
}

func Fingerprint(key string) []byte {
	entity, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(key))
	log.Check(log.WarnLevel, "Reading user public key", err)

	for _, v := range entity {
		return v.PrimaryKey.Fingerprint[:]
	}
	return []byte("")
}

func SignHub(owner, hash string) string {
	//Here some authorization to HUB should be added

	resp, err := http.Get("https://hub.subut.ai/some/e2e/endpoint?owner" + owner + "&hash=" + hash)
	if log.Check(log.WarnLevel, "Sending sign request", err) || resp.StatusCode != 200 {
		return ""
	}
	signature, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if log.Check(log.WarnLevel, "Reading response body", err) {
		return ""
	}
	return string(signature)
}
