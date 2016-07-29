package pgp

import (
	"bytes"
	// "io/ioutil"
	// "net/http"

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

//Until we get in coordination with Hub's devs this is only theoretical dummy function
func SignHub(owner, hash string) string {
	//Here some authorization to HUB should be added

	// resp, err := http.Get("https://hub.subut.ai/some/e2e/endpoint?owner" + owner + "&hash=" + hash)
	// if log.Check(log.WarnLevel, "Sending sign request", err) || resp.StatusCode != 200 {
	// return ""
	// }
	// signature, err := ioutil.ReadAll(resp.Body)
	// resp.Body.Close()
	// if log.Check(log.WarnLevel, "Reading response body", err) {
	// return ""
	// }
	// return string(signature)
	return `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA1

e3fc50a88d0a364313df4b21ef20c29e
-----BEGIN PGP SIGNATURE-----
Version: GnuPG v1

iQEcBAEBAgAGBQJXmzTXAAoJEMke5E6pQEpU1MYH+gIRKzDxYBtE6v4gI/eifjT2
tHEyNZEF1Oi7fIngtj7ZKnVYC7Yfgjxi7+49MbJchGtHDP2pYQNBo+aAUgGaRShq
DT5/Xnnx9K3gVY5lGzDdHCHUI+uICjaHrg0LpG+CbIoSNy51Jzmey2s4yTPkxg0u
lKboTv4/k5BBFGxRdGhT9AVKDFurcmHwCkcCMr8eQ3hO9+Gvc4UVN8mGM0i7tR/4
EmZhKgbtgc1IQcPqAbDaXtFPuZFLo+CJPBTLQWEHqsGlB9GnDy4sf0AV1Wox69oN
Mi5hnC7esCsXzC4ZBwegpIyJmXTKT8+2ErZoqrCM9H73UP+C3LGof5AJ+SoFuzo=
=DETD
-----END PGP SIGNATURE-----
`
}
