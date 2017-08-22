package pgp

import (
	"bytes"

	"github.com/subutai-io/agent/log"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"

	"github.com/subutai-io/gorjun/db"
)

func Verify(name, message string) string {
	for _, key := range db.UserKeys(name) {
		entity, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(key))
		log.Check(log.WarnLevel, "Reading user public key", err)

		if block, _ := clearsign.Decode([]byte(message)); block != nil {
			_, err = openpgp.CheckDetachedSignature(entity, bytes.NewBuffer(block.Bytes), block.ArmoredSignature.Body)
			if !log.Check(log.WarnLevel, "Checking signature", err) {
				return string(block.Bytes)
			}
		}
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
