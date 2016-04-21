package auth

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/pgp"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(`<html><title>Registration</title><body>
			<form action="/auth/register" method="post">
				Name: <input type="text" name="name"><br>
				PGP public key: <textarea cols="63" rows="30" name="key"></textarea><br>
				<input type="submit" name="submit" value="Submit">
			</form></body></html>`))
	} else if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		name := r.MultipartForm.Value["name"][0]
		key := r.MultipartForm.Value["key"][0]
		w.Write([]byte("Name: " + name + "\n"))
		w.Write([]byte("PGP key: " + key + "\n"))
		db.RegisterUser([]byte(name), []byte(key))
	}
}

func Token(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		name := r.URL.Query().Get("name")
		hash := md5.New()
		hash.Write([]byte(time.Now().String() + name))
		token := fmt.Sprintf("%x", hash.Sum(nil))
		db.SaveAuthID(name, token)
		w.Write([]byte(token))
	} else if r.Method == http.MethodPost {
		r.ParseMultipartForm(32 << 20)
		name := r.MultipartForm.Value["name"][0]
		message := r.MultipartForm.Value["message"][0]
		authid := pgp.Verify(name, message)
		if db.CheckAuthID(authid) == name {
			hash := md5.New()
			hash.Write([]byte(time.Now().String() + name))
			token := fmt.Sprintf("%x", hash.Sum(nil))
			db.SaveToken(name, token)
			w.Write([]byte(token))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Signature verification failed"))
		}
	}

}
