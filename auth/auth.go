package auth

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/pgp"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if strings.Split(r.RemoteAddr, ":")[0] == "127.0.0.1" && r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		if len(r.MultipartForm.Value["name"]) > 0 && len(r.MultipartForm.Value["key"]) > 0 {
			name := r.MultipartForm.Value["name"][0]
			key := r.MultipartForm.Value["key"][0]

			w.Write([]byte("Name: " + name + "\n"))
			w.Write([]byte("PGP key: " + key + "\n"))

			db.RegisterUser([]byte(name), []byte(key))
			return
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Not allowed"))
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
