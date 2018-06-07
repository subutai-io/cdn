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
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
)

type share struct {
	Token  string   `json:"token"`
	Id     string   `json:"id"`
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
	Repo   string   `json:"repo"`
}

// Handler function works with income upload requests, makes sanity checks, etc
func Handler(w http.ResponseWriter, r *http.Request) (md5sum, sha256sum, owner string) {
	token := strings.ToLower(r.Header.Get("token"))
	owner = strings.ToLower(db.TokenOwner(token))
	log.Debug(fmt.Sprintf("Upload request: %+v"))
	log.Debug(fmt.Sprintf("token: %+v, owner: %+v", token, owner))
	if len(token) == 0 || len(owner) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized upload request")
		return
	}
	repo := strings.Split(r.URL.EscapedPath(), "/")
	log.Debug(fmt.Sprintf("repo: %+v", repo))
	if len(repo) < 4 {
		log.Warn(r.URL.EscapedPath() + " - bad deletion request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
		return
	}
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	log.Info(fmt.Sprintf("file: %T, header.Filename: %s", file, header.Filename))
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
	out, err := os.Create(config.ConfigurationStorage.Path + header.Filename)
	log.Debug(fmt.Sprintf("Creating file for writing: %+v, %+v", out, err))
	if log.Check(log.WarnLevel, "Unable to create the file for writing", err) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot create file"))
		return
	}
	defer out.Close()
	limit := int64(db.QuotaLeft(owner))
	log.Debug(fmt.Sprintf("limit left: %+v", limit))
	f := io.Reader(file)
	if limit != -1 {
		f = io.LimitReader(file, limit)
	}
	// write the content from POST to the file
	if copied, err := io.Copy(out, f); limit != -1 && (copied == limit || err != nil) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write file or storage quota exceeded"))
		log.Warn("User " + owner + " exceeded storage quota, removing file")
		os.Remove(config.ConfigurationStorage.Path + header.Filename)
		return
	} else {
		db.QuotaUsageSet(owner, int(copied))
		log.Info("User " + owner + ", quota usage +" + strconv.Itoa(int(copied)))
	}
	md5sum = Hash(config.ConfigurationStorage.Path + header.Filename)
	sha256sum = Hash(config.ConfigurationStorage.Path + header.Filename, "sha256")
	if len(md5sum) == 0 || len(sha256sum) == 0 {
		log.Warn("Failed to calculate hash for " + header.Filename)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to calculate hash"))
		return
	}
	if repo[3] != "apt" {
		log.Debug(fmt.Sprintf("repo[3] is not apt. Renaming %+v to %+v", config.ConfigurationStorage.Path + header.Filename, config.ConfigurationStorage.Path + md5sum))
		os.Rename(config.ConfigurationStorage.Path + header.Filename, config.ConfigurationStorage.Path + md5sum)
	}
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
	token := strings.ToLower(r.URL.Query().Get("token"))
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Empty file id"))
		log.Warn(r.RemoteAddr + " - empty file id")
		return ""
	}
	user := db.TokenOwner(token)
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
	if db.CheckRepo(user, []string{repo[3]}, id) == 0 && user != "subutai" {
		log.Warn("File " + info["name"] + "(" + id + ") in " + repo[3] + " repo is not owned by " + user + ", rejecting deletion request")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("File " + info["name"] + " not found or it has different owner"))
		return ""
	}
	user = db.FileField(id, "owner")[0]

	md5, _ := db.Hash(id)
	f, err := os.Stat(config.ConfigurationStorage.Path + md5)
	if !log.Check(log.WarnLevel, "Reading file stats", err) {
		db.QuotaUsageSet(user, -int(f.Size()))
		log.Info("User " + user + ", quota usage -" + strconv.Itoa(int(f.Size())))
	}
	db.Delete(user, repo[3], id)
	if db.CountMd5(md5) == 0 && repo[3] != "apt" {
		log.Warn("Removing " + id + " from disk")
		// torrent.Delete(id)
		if log.Check(log.WarnLevel, "Removing "+info["name"]+"from disk", os.Remove(config.ConfigurationStorage.Path+md5)) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to remove file"))
			return ""
		}
	}

	if repo[3] == "apt" && db.CountMd5(info["md5"]) == 0 {
		if log.Check(log.WarnLevel, "Removing "+info["name"]+"from disk", os.Remove(config.ConfigurationStorage.Path+info["Filename"])) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to remove file"))
			return ""
		}
	}

	log.Info("Removing " + info["name"] + " from " + repo[3] + " repo")
	return id
}

// Share receives HTTP Request of type application/json and handles it
func Share(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Debug(fmt.Sprintf("POST Share request: %+v", r))
		log.Debug(fmt.Sprintf("Body: %+v", r.Body))
		var err error
		var data share
		r.ParseMultipartForm(32 << 20)
		formValue := r.FormValue("json")
		log.Debug(fmt.Sprintf("r.FormValue(\"json\"): %+v", formValue))
		multipartFormValue := make([]string, 0)
		if r.MultipartForm != nil {
			multipartFormValue = r.MultipartForm.Value["json"]
			log.Debug(fmt.Sprintf("r.MultipartForm: %+v", multipartFormValue))
		}
		log.Debug(fmt.Sprintf("form: %+v", formValue))
		log.Debug(fmt.Sprintf("multipart: %+v", multipartFormValue))
		if len(formValue) == 0 && len(multipartFormValue) == 0 {
			log.Debug(fmt.Sprintf("Both empty. Body: %+v", r.Body))
			b := r.Body
			d := json.NewDecoder(b)
			if err = d.Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Empty json"))
				log.Warn("Share request: empty json, nothing to do")
				return
			}
		} else {
			err = json.Unmarshal([]byte(formValue), &data)
			if err != nil {
				err = json.Unmarshal([]byte(multipartFormValue[0]), &data)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Failed to parse json body"))
					return
				}
			}
		}
		if len(data.Token) == 0 || len(db.TokenOwner(data.Token)) == 0 {
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
		owner := strings.ToLower(db.TokenOwner(data.Token))
		if db.CheckRepo(owner, []string{data.Repo}, data.Id) == 0 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("File is not owned by authorized user"))
			log.Warn("User tried to share another's file, rejecting")
			return
		}
		for _, v := range data.Add {
			log.Info("Sharing " + data.Id + " with " + v)
			db.AddShare(data.Id, owner, v)
		}
		for _, v := range data.Remove {
			log.Info("Unsharing " + data.Id + " with " + v)
			db.RemoveShare(data.Id, owner, v)
		}
	} else if r.Method == "GET" {
		id := r.URL.Query().Get("id")
		if len(id) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Empty file id"))
			return
		}
		token := strings.ToLower(r.URL.Query().Get("token"))
		if len(token) == 0 || len(db.TokenOwner(token)) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not authorized"))
			return
		}
		owner := db.TokenOwner(token)
		repo := r.URL.Query().Get("repo")
		if len(repo) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Repository not specified"))
			return
		}
		if db.CheckRepo(owner, []string{repo}, id) == 0 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("File is not owned by authorized user"))
			log.Warn("User tried to request scope of another's file, rejecting")
			return
		}
		js, _ := json.Marshal(db.GetFileScope(id, strings.ToLower(owner)))
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
		token := strings.ToLower(r.URL.Query().Get("token"))

		if len(token) == 0 || len(db.TokenOwner(token)) == 0 || db.TokenOwner(token) != "Hub" && !strings.EqualFold(db.TokenOwner(token), user) {
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

		if len(token) == 0 || len(db.TokenOwner(token)) == 0 || db.TokenOwner(token) != "Hub" && db.TokenOwner(token) != "subutai" {
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
