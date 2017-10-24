package upload

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
)

type share struct {
	Token  string   `json:"token"`
	Id     string   `json:"id"`
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
	Repo   string   `json:"repo"`
}

//Handler function works with income upload requests, makes sanity checks, etc
func Handler(w http.ResponseWriter, r *http.Request) (md5sum, sha256sum, owner string) {
	r.ParseMultipartForm(32 << 20)
	if len(r.MultipartForm.Value["token"]) == 0 || len(db.CheckToken(r.MultipartForm.Value["token"][0])) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized upload request")
		return
	}

	owner = strings.ToLower(db.CheckToken(r.MultipartForm.Value["token"][0]))

	file, header, err := r.FormFile("file")
	if log.Check(log.WarnLevel, "Failed to parse POST form", err) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot get file from request"))
		return
	}
	defer file.Close()

	if !сheckLength(owner, r.Header.Get("Content-Length")) {
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
	} else {
		db.QuotaUsageSet(owner, int(copied))
		log.Info("User " + owner + ", quota usage +" + strconv.Itoa(int(copied)))
	}

	md5sum = Hash(config.Storage.Path + header.Filename)
	sha256sum = Hash(config.Storage.Path+header.Filename, "sha256")
	if len(md5sum) == 0 || len(sha256sum) == 0 {
		log.Warn("Failed to calculate hash for " + header.Filename)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to calculate hash"))
		return
	}

	if len(r.MultipartForm.Value["private"]) > 0 && r.MultipartForm.Value["private"][0] == "true" {
		log.Info("Sharing " + md5sum + " with " + owner)
		db.ShareWith(md5sum, owner, owner)
	}

	os.Rename(config.Storage.Path+header.Filename, config.Storage.Path+md5sum)
	log.Info("File received: " + header.Filename + "(" + md5sum + ")")

	return md5sum, sha256sum, owner
}

func Hash(file string, algo ...string) string {
	f, err := os.Open(file)
	log.Check(log.WarnLevel, "Opening file "+file, err)
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
	id := r.URL.Query().Get("id")
	token := r.URL.Query().Get("token")
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty file id"))
		log.Warn(r.RemoteAddr + " - empty file id")
		return ""
	}
	user := db.CheckToken(token)
	if len(token) == 0 || len(user) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Failed to authorize using provided token"))
		log.Warn(r.RemoteAddr + " - Failed to authorize using provided token")
		return ""
	}
	info := db.Info(id)
	if len(info) == 0 {
		log.Warn("File not found by id")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("File not found"))
		return ""
	}

	repo := strings.Split(r.URL.EscapedPath(), "/")
	if len(repo) < 4 {
		log.Warn(r.URL.EscapedPath() + " - bad deletion request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
		return ""
	}

	if db.CheckRepo(user, repo[3], id) == 0 {
		log.Warn("File " + info["name"] + "(" + id + ") in " + repo[3] + " repo is not owned by " + user + ", rejecting deletion request")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("File " + info["name"] + " not found or it has different owner"))
		return ""
	}
	md5, _ := db.Hash(id)
	f, err := os.Stat(config.Storage.Path + md5)
	if !log.Check(log.WarnLevel, "Reading file stats", err) {
		db.QuotaUsageSet(user, -int(f.Size()))
		log.Info("User " + user + ", quota usage -" + strconv.Itoa(int(f.Size())))
	}

	if db.Delete(user, repo[3], id) == 0 {
		log.Warn("Removing " + id + " from disk")
		// torrent.Delete(id)
		if log.Check(log.WarnLevel, "Removing "+info["name"]+"from disk", os.Remove(config.Storage.Path+md5)) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to remove file"))
			return ""
		}
	}

	log.Info("Removing " + info["name"] + " from " + repo[3] + " repo")
	return id
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
		if len(data.Repo) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Empty repo name"))
			log.Warn("Empty repo name, rejecting share request")
			return
		}
		owner := db.CheckToken(data.Token)
		if db.CheckRepo(owner, data.Repo, data.Id) == 0 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("File is not owned by authorized user"))
			log.Warn("User tried to share another's file, rejecting")
			return
		}
		for _, v := range data.Add {
			log.Info("Sharing " + data.Id + " with " + v)
			db.ShareWith(data.Id, owner, v)
		}
		for _, v := range data.Remove {
			log.Info("Unsharing " + data.Id + " with " + v)
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
		repo := r.URL.Query().Get("repo")
		if len(repo) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Repository not specified"))
			return
		}
		if db.CheckRepo(owner, repo, id) == 0 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("File is not owned by authorized user"))
			log.Warn("User tried to request scope of another's file, rejecting")
			return
		}
		js, _ := json.Marshal(db.GetScope(id, strings.ToLower(owner)))
		w.Write(js)
	}
}

func сheckLength(user, length string) bool {
	l, err := strconv.Atoi(length)
	if err != nil || len(length) == 0 || l < db.QuotaLeft(user) || db.QuotaLeft(user) == -1 {
		return true
	}
	return false
}

func Quota(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user := r.URL.Query().Get("user")
		fix := r.URL.Query().Get("fix")
		token := r.URL.Query().Get("token")

		if len(token) == 0 || len(db.CheckToken(token)) == 0 || db.CheckToken(token) != "Hub" && !strings.EqualFold(db.CheckToken(token), user) {
			w.Write([]byte("Forbidden"))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if len(user) != 0 {
			user = strings.ToLower(user)
			q, _ := json.Marshal(map[string]int{
				"quota": db.QuotaGet(user),
				"used":  db.QuotaUsageGet(user),
				"left":  db.QuotaLeft(user)})
			w.Write([]byte(q))
		}
		if user == "subutai" && len(fix) != 0 {
			db.QuotaUsageCorrect()
		}

	} else if r.Method == "POST" {
		user := r.FormValue("user")
		quota := r.FormValue("quota")
		token := r.FormValue("token")

		if len(token) == 0 || len(db.CheckToken(token)) == 0 || db.CheckToken(token) != "Hub" && db.CheckToken(token) != "subutai" {
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
