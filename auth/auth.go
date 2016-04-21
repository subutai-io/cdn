package auth

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/optdyn/gorjun/db"
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
		w.Write([]byte("Name: " + r.PostFormValue("name") + "\n"))
		w.Write([]byte("PGP key: " + r.PostFormValue("key") + "\n"))
		db.RegisterUser([]byte(r.PostFormValue("name")), []byte(r.PostFormValue("key")))
	}
}

func Token(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	hash := md5.New()
	hash.Write([]byte(time.Now().String() + name))
	token := fmt.Sprintf("%x", hash.Sum(nil))
	db.SaveToken(name, token)
	w.Write([]byte("user: " + db.CheckToken(token) + " token: " + token))
}
