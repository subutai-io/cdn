package upload

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/optdyn/gorjun/db"

	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

func Page(repo string) string {
	return `
  <html>
  <title>Go upload</title>
  <body>
  <form action="/kurjun/rest/` + repo + `/upload" method="post" enctype="multipart/form-data">
  <label for="file">Filename:</label>
  <input type="file" name="file" id="file">
  <input type="submit" name="submit" value="Submit">
  </form>
  </body>
  </html>
`
}

func Handler(w http.ResponseWriter, r *http.Request) string {
	r.ParseMultipartForm(32 << 20)
	if len(r.MultipartForm.Value["token"]) == 0 || len(db.CheckToken(r.MultipartForm.Value["token"][0])) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting not authorized upload request")
		return ""
	}
	token := r.MultipartForm.Value["token"][0]
	log.Info("User: " + r.RemoteAddr + " token: " + token)
	file, header, err := r.FormFile("file")
	if log.Check(log.WarnLevel, "Failed to parse POST form", err) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot get file from request"))
		return ""
	}
	defer file.Close()
	out, err := os.Create(path + header.Filename)
	if log.Check(log.WarnLevel, "Unable to create the file for writing", err) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cannot create file"))
		return ""
	}
	defer out.Close()
	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if log.Check(log.WarnLevel, "Writing file", err) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write file"))
		return ""
	}
	hash := genHash(path + header.Filename)
	if len(hash) == 0 {
		log.Warn("Failed to calculate hash for " + header.Filename)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to calculate hash"))
		return ""
	}
	os.Rename(path+header.Filename, path+hash)
	log.Info("File uploaded successfully: " + header.Filename + "(" + hash + ")")
	return hash
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