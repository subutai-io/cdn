package upload

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"gorjun/torrent"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
)

type share struct {
	Token  string   `json:"token"`
	Id     string   `json:"id"`
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

//Handler function works with income upload requests, makes sanity checks, etc
func Handler(w http.ResponseWriter, r *http.Request) (hash, owner string) {
	r.ParseMultipartForm(32 << 20)
	if len(r.MultipartForm.Value["token"]) == 0 || len(db.CheckToken(r.MultipartForm.Value["token"][0])) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized upload request")
		return
	}

	owner = db.CheckToken(r.MultipartForm.Value["token"][0])

	file, header, err := r.FormFile("file")
	defer file.Close()
	if log.Check(log.WarnLevel, "Failed to parse POST form", err) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot get file from request"))
		return
	}

	l := сheckLength(owner, r.Header.Get("Content-Length"))
	if !l {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Storage quota exceeded"))
		log.Warn("User " + owner + " exceeded storage quota, rejecting upload")
		return
	}

	out, err := os.Create(config.Storage.Path + header.Filename)
	if log.Check(log.WarnLevel, "Unable to create the file for writing", err) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot create file"))
		return
	}
	defer out.Close()

	limit := int64(db.QuotaLeft(owner))
	f := io.Reader(file)
	if limit != -1 {
		f = io.LimitReader(file, limit)
	}

	// write the content from POST to the file
	if copied, err := io.Copy(out, f); limit != -1 && (copied == limit || err != nil) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write file or storage quota exceeded"))
		log.Warn("User " + owner + " exceeded storage quota, removing file")
		os.Remove(config.Storage.Path + header.Filename)
		return
	}

	hash = Hash(config.Storage.Path + header.Filename)
	if len(hash) == 0 {
		log.Warn("Failed to calculate hash for " + header.Filename)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to calculate hash"))
		return
	}

	os.Rename(config.Storage.Path+header.Filename, config.Storage.Path+hash)
	log.Info("File uploaded successfully: " + header.Filename + "(" + hash + ")")

	return hash, owner
}

func Hash(file string, algo ...string) string {
	f, err := os.Open(file)
	log.Check(log.WarnLevel, "Opening file"+file, err)
	defer f.Close()

	hash := md5.New()
	if len(algo) != 0 {
		switch algo[0] {
		case "sha512":
			hash = sha512.New()
		case "sha256":
			hash = sha256.New()
		case "sha1":
			hash = sha1.New()
		}
	}
	if _, err := io.Copy(hash, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Delete(w http.ResponseWriter, r *http.Request) string {
	hash := r.URL.Query().Get("id")
	token := r.URL.Query().Get("token")
	if len(hash) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty file id"))
		log.Warn(r.RemoteAddr + " - empty file id")
		return ""
	}
	if len(token) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Empty token"))
		log.Warn(r.RemoteAddr + " - empty token")
		return ""
	}
	user := db.CheckToken(token)
	info := db.Info(hash)
	if len(info) == 0 {
		log.Warn("File not found by hash")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("File not found"))
		return ""
	}

	if !db.CheckOwner(user, hash) {
		log.Warn("File " + info["name"] + "(" + hash + ") is not owned by " + user + ", rejecting deletion request")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("File " + info["name"] + " is not owned by " + user))
		return ""
	}

	f, _ := os.Stat(config.Storage.Path + hash)
	db.QuotaUsageSet(user, -int(f.Size()))

	if db.Delete(user, hash) <= 0 {
		torrent.Delete(hash)
		if log.Check(log.WarnLevel, "Removing "+info["name"]+"from disk", os.Remove(config.Storage.Path+hash)) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to remove file"))
			return ""
		}
	}

	log.Info("Removing " + info["name"])
	return hash
}

func Share(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if len(r.FormValue("json")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Empty json"))
			log.Warn("Share request: empty json, nothing to do")
			return
		}
		var data share
		if log.Check(log.WarnLevel, "Parsing share request json", json.Unmarshal([]byte(r.FormValue("json")), &data)) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to parse json body"))
			return
		}
		if len(data.Token) == 0 || len(db.CheckToken(data.Token)) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not authorized"))
			log.Warn("Empty or invalid token, rejecting share request")
			return
		}
		if len(data.Id) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Empty file id"))
			log.Warn("Empty file id, rejecting share request")
			return
		}
		owner := db.CheckToken(data.Token)
		if !db.CheckOwner(owner, data.Id) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("File is not owned by authorized user"))
			log.Warn("User tried to share another's file, rejecting")
			return
		}
		for _, v := range data.Add {
			db.ShareWith(data.Id, owner, v)
		}
		for _, v := range data.Remove {
			db.UnshareWith(data.Id, owner, v)
		}
	} else if r.Method == "GET" {
		id := r.URL.Query().Get("id")
		if len(id) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Empty file id"))
			return
		}
		token := r.URL.Query().Get("token")
		if len(token) == 0 || len(db.CheckToken(token)) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not authorized"))
			return
		}
		owner := db.CheckToken(token)
		if !db.CheckOwner(owner, id) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("File is not owned by authorized user"))
			log.Warn("User tried to request scope of another's file, rejecting")
			return
		}
		js, _ := json.Marshal(db.GetScope(id, owner))
		w.Write(js)
	}
}

func сheckLength(user, length string) bool {
	l, err := strconv.Atoi(length)
	if err != nil || len(length) == 0 {
		log.Warn("Empty or invalid content length")
		return true
	}

	if l < db.QuotaLeft(user) || db.QuotaLeft(user) == -1 {
		db.QuotaUsageSet(user, l)
		return true
	}
	return false
}

func Quota(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user := r.URL.Query().Get("user")
		token := r.URL.Query().Get("token")

		if len(token) == 0 || len(db.CheckToken(token)) == 0 || db.CheckToken(token) != "Hub" && db.CheckToken(token) != "2129bb4fb65b27ff68a21c678d461db2e2c20bb7" && db.CheckToken(token) != user {
			w.Write([]byte("Forbidden"))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if len(user) != 0 {
			q, _ := json.Marshal(map[string]int{
				"quota": db.QuotaGet(user),
				"used":  db.QuotaUsageGet(user),
				"left":  db.QuotaLeft(user)})
			w.Write([]byte(q))
		}

	} else if r.Method == "POST" {
		user := r.FormValue("user")
		quota := r.FormValue("quota")
		token := r.FormValue("token")

		if len(token) == 0 || len(db.CheckToken(token)) == 0 || db.CheckToken(token) != "Hub" && db.CheckToken(token) != "2129bb4fb65b27ff68a21c678d461db2e2c20bb7" {
			w.Write([]byte("Forbidden"))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if len(user) == 0 || len(quota) == 0 {
			w.Write([]byte("Please specify username and quota value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if q, err := strconv.Atoi(quota); err != nil || q < -1 {
			w.Write([]byte("Invalid quota value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		db.QuotaSet(user, quota)
		log.Info("New quota for " + user + " is " + quota)
		w.Write([]byte("Ok"))
		w.WriteHeader(http.StatusOK)
	}
}
