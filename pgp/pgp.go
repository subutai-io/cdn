package pgp

import (
	"bytes"
	"net/http"

	"github.com/optdyn/gorjun/db"

	"github.com/subutai-io/base/agent/log"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

func Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(`<html><title>Authorization</title><body>
			<form action="/pgp/verify" method="post">
				Name: <input type="text" name="name"><br>
				Signed message: <textarea cols="63" rows="30" name="message"></textarea><br>
				<input type="submit" name="submit" value="Submit">
			</form></body></html>`))
	} else if r.Method == "POST" {
		user := r.PostFormValue("name")
		message := r.PostFormValue("message")
		key := db.UserKey(user)

		entity, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(key))
		log.Check(log.WarnLevel, "Reading user public key", err)

		block, _ := clearsign.Decode([]byte(message))

		_, err = openpgp.CheckDetachedSignature(entity, bytes.NewBuffer(block.Bytes), block.ArmoredSignature.Body)
		if log.Check(log.WarnLevel, "Checking signature", err) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Signature verification failed"))
			return
		}

		w.Write([]byte("User: " + user + "\n"))
		w.Write(block.Bytes)
	}
}
