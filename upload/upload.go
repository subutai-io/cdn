package upload

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
)

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

	out, err := os.Create(config.Filepath + header.Filename)
	defer out.Close()
	if log.Check(log.WarnLevel, "Unable to create the file for writing", err) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot create file"))
		return
	}

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if log.Check(log.WarnLevel, "Writing file", err) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write file"))
		return
	}

	hash = genHash(config.Filepath + header.Filename)
	if len(hash) == 0 {
		log.Warn("Failed to calculate hash for " + header.Filename)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to calculate hash"))
		return
	}

	os.Rename(config.Filepath+header.Filename, config.Filepath+hash)
	log.Info("File uploaded successfully: " + header.Filename + "(" + hash + ")")
	// log.Info("Signature: " + signature)
	return hash, owner
}

func genHash(file string) string {
	f, err := os.Open(file)
	log.Check(log.WarnLevel, "Opening file"+file, err)
	defer f.Close()

	hash := md5.New()
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
		log.Warn("File " + info["name"] + "(" + hash + ") is not owned by " + user + ", rejecting")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("File " + info["name"] + " is not owned by " + user))
		return ""
	}
	if db.Delete(user, hash) <= 0 {
		if log.Check(log.WarnLevel, "Removing "+info["name"]+"from disk", os.Remove(config.Filepath+hash)) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to remove file"))
			return ""
		}
	}
	log.Info("Removing " + info["name"])
	return hash
}
