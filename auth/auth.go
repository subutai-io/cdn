package auth

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/subutai-io/agent/log"

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
		name := r.FormValue("user")
		message := r.FormValue("message")
		if len(name) == 0 || len(message) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please specify user name and auth message"))
			log.Warn(r.RemoteAddr + " - empty user name or message filed")
			return
		}
		authid := pgp.Verify(name, message)
		if db.CheckAuthID(authid) == name {
			token := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(time.Now().String(), name, rand.Float64()))))
			db.SaveToken(name, fmt.Sprintf("%x", sha256.Sum256([]byte(token))))
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

func Key(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if len(user) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty user"))
		return
	}
	key := db.UserKey(user)
	if len(key) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User key not found"))
		return
	}
	w.Write([]byte(key))
}

func Sign(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	if len(r.MultipartForm.Value["token"]) == 0 || len(db.CheckToken(r.MultipartForm.Value["token"][0])) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized sign request")
		return
	}
	owner := db.CheckToken(r.MultipartForm.Value["token"][0])
	if len(r.MultipartForm.Value["signature"]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty signature"))
		log.Warn("auth.Sign received empty signature")
		return
	}
	signature := r.MultipartForm.Value["signature"][0]
	hash := pgp.Verify(owner, signature)
	if len(hash) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Failed to verify signature with user key"))
		log.Warn("Failed to verify signature with user key")
		return
	}
	if db.CheckRepo(owner, "", hash) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("File and signature have different owner"))
		log.Warn("File and signature have different owner")
		return
	}
	db.Write(owner, hash, "", map[string]string{"signature": signature})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File " + hash + " has been signed by " + owner))
	log.Info("File " + hash + " has been signed by " + owner)
	return
}
