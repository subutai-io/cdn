package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/cdn/pgp"
	"strings"
	"math/rand"
	"time"
	"crypto/md5"
	"crypto/sha256"
)

type Hashes struct {
	Md5    string `json:"md5,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

type OldResult struct {
	FileID        string   `json:"id,omitempty"`
	Owner         []string `json:"owner,omitempty"`
	Name          string   `json:"name,omitempty"`
	Filename      string   `json:"filename,omitempty"`
	Version       string   `json:"version,omitempty"`
	Hash          Hashes   `json:"hash,omitempty"`
	Size          int64    `json:"size,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Date          string   `json:"upload-date-formatted,omitempty"`
	Timestamp     string   `json:"upload-date-timestamp,omitempty"`
	Description   string   `json:"description,omitempty"`
	Architecture  string   `json:"architecture,omitempty"`
	Parent        string   `json:"parent,omitempty"`
	ParentVersion string   `json:"parent-version,omitempty"`
	ParentOwner   string   `json:"parent-owner,omitempty"`
	PrefSize      string   `json:"prefsize,omitempty"`
}

// FileSearch handles the info and list HTTP requests
func FileSearch(w http.ResponseWriter, r *http.Request) {
	log.Info("Received FileSearch request")
	log.Info(r)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method for info/list request"))
		return
	}
	request := new(SearchRequest)
	request.InitValidators()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect info/list request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	files := request.Retrieve()
	log.Info("Successfully retrieved files")
	oldFiles := make([]*OldResult, 0)
	for _, file := range files {
		oldFiles = append(oldFiles, file.ConvertToOld())
	}
	log.Info("Retrieve: ", oldFiles)
	log.Info("Successfully converted files to old format")
	result, _ := json.Marshal(oldFiles)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
	log.Info("Successfully handled FileSearch request")
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	log.Info("Received upload request")
	log.Info(r)
	if r.Method != "POST" {
		log.Warn("Incorrect method for upload request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect method for upload request")))
		return
	}
	request := new(UploadRequest)
	request.InitUploaders()
	log.Info("Successfully initialized request")
	err := request.ParseRequest(r)
	if err != nil {
		log.Warn("Couldn't parse upload request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Incorrect upload request: %v", err)))
		return
	}
	log.Info("Successfully parsed request")
	err = request.Upload()
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't upload file: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while uploading file: %v", err)))
		return
	}
	log.Info("Successfully uploaded a file: ", request.fileID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(request.fileID))
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		if strings.Split(r.RemoteAddr, ":")[0] == "127.0.0.1" && len(r.MultipartForm.Value["name"]) > 0 && len(r.MultipartForm.Value["key"]) > 0 {
			name := r.MultipartForm.Value["name"][0]
			key := r.MultipartForm.Value["key"][0]
			w.Write([]byte("Name: " + name + "\n"))
			w.Write([]byte("PGP key: " + key + "\n"))
			db.RegisterUser([]byte(name), []byte(key))
			log.Info("User " + name + " registered with this key " + key)
			return
		} else if len(r.MultipartForm.Value["key"]) > 0 {
			key := pgp.Verify("Hub", r.MultipartForm.Value["key"][0])
			log.Debug(fmt.Sprintf("Key == %+v", r.MultipartForm.Value["key"]))
			if len(key) == 0 {
				log.Debug(fmt.Sprintf("Key empty"))
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
			if len(r.MultipartForm.Value["name"]) > 0 {
				db.RegisterUser([]byte(r.MultipartForm.Value["name"][0]), []byte(key))
			} else {
				db.RegisterUser([]byte(fmt.Sprintf("%x", fingerprint)), []byte(key))
			}
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
		log.Info(fmt.Sprintf("Received token POST: name: %+v, message: %+v", name, message))
		if len(name) == 0 || len(message) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please specify user name and auth message"))
			log.Warn(r.RemoteAddr + " - empty user name or message filed")
			return
		}
		authid := pgp.Verify(name, message)
		log.Info(fmt.Sprintf("Checking AuthID %s", authid))
		if authIDName := db.CheckAuthID(authid); authIDName == name {
			token := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(time.Now().String(), name, rand.Float64()))))
			db.SaveToken(name, fmt.Sprintf("%x", sha256.Sum256([]byte(token))))
			w.Write([]byte(token))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("Signature verification failed: %s", authIDName)))
		}
	}
}

func Validate(w http.ResponseWriter, r *http.Request) {
	token := strings.ToLower(r.URL.Query().Get("token"))
	if len(token) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty token"))
		return
	}
	if len(db.TokenOwner(token)) == 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}
	w.Write([]byte("Success"))
}

// Keys returns list of users GPG keys
func Keys(w http.ResponseWriter, r *http.Request) {
	if keys := db.UserKeys(r.URL.Query().Get("user")); len(keys) > 0 {
		if out, err := json.Marshal(keys); err == nil {
			w.Write(out)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// Key is replaced by Keys and left for compatibility. This function should be removed later.
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
	if len(r.MultipartForm.Value["token"]) == 0 || len(db.TokenOwner(r.MultipartForm.Value["token"][0])) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized sign request")
		return
	}
	owner := db.TokenOwner(r.MultipartForm.Value["token"][0])
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
	if /* db.CheckShare(hash, owner) */ db.CheckRepo(owner, []string{}, hash) == 0 {
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

func Owner(w http.ResponseWriter, r *http.Request) {
	token := strings.ToLower(r.URL.Query().Get("token"))
	owner := strings.ToLower(db.TokenOwner(token))
	if len(token) == 0 || len(owner) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized owner request")
		return
	}
	w.Write([]byte(owner))
	return
}
