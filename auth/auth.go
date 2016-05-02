package auth

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/pgp"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		if strings.Split(r.RemoteAddr, ":")[0] == "127.0.0.1" && len(r.MultipartForm.Value["name"]) > 0 && len(r.MultipartForm.Value["key"]) > 0 {
			name := r.MultipartForm.Value["name"][0]
			key := r.MultipartForm.Value["key"][0]

			w.Write([]byte("Name: " + name + "\n"))
			w.Write([]byte("PGP key: " + key + "\n"))

			db.RegisterUser([]byte(name), []byte(key))
			return
		} else if len(r.MultipartForm.Value["key"]) > 0 {
			key := pgp.Verify("Hub", r.MultipartForm.Value["key"][0])
			if len(key) == 0 {
				w.Write([]byte("Signature check failed"))
				w.WriteHeader(http.StatusForbidden)
				return
			}
			fingerprint := pgp.Fingerprint(key)
			if len(fingerprint) == 0 {
				w.Write([]byte("Filed to get key fingerprint"))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			db.RegisterUser([]byte(fmt.Sprintf("%x", fingerprint)), []byte(key))
			return
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Not allowed"))
}

func Token(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	if r.Method == http.MethodGet {
		name := r.URL.Query().Get("user")
		if len(name) != 0 {
			hash := md5.New()
			hash.Write([]byte(fmt.Sprint(time.Now().String(), name, rand.Float64())))
			authID := fmt.Sprintf("%x", hash.Sum(nil))
			db.SaveAuthID(name, authID)
			w.Write([]byte(authID))
		}
	} else if r.Method == http.MethodPost {
		r.ParseMultipartForm(32 << 20)
		name := r.MultipartForm.Value["user"][0]
		message := r.MultipartForm.Value["message"][0]
		authid := pgp.Verify(name, message)
		if db.CheckAuthID(authid) == name {
			hash := sha256.New()
			hash.Write([]byte(fmt.Sprint(time.Now().String(), name, rand.Float64())))
			token := fmt.Sprintf("%x", hash.Sum(nil))
			hash.Reset()
			hash.Write([]byte(token))

			db.SaveToken(name, fmt.Sprintf("%x", hash.Sum(nil)))
			w.Write([]byte(token))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Signature verification failed"))
		}
	}
}

func Validate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty token"))
		return
	}
	if len(db.CheckToken(token)) == 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}
	w.Write([]byte("Success"))
}
