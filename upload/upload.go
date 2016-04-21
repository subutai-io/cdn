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
  <form action="/` + repo + `/upload" method="post" enctype="multipart/form-data">
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
	token := r.MultipartForm.Value["token"][0]
	fmt.Println("token: ", token)
	if len(token) == 0 || len(db.CheckToken(token)) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid token"))
		return ""
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Warn(err.Error())
		return ""
	}
	defer file.Close()
	out, err := os.Create(path + header.Filename)
	if err != nil {
		log.Warn("Unable to create the file for writing. Check your write access privilege")
		return ""
	}
	defer out.Close()
	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		log.Warn(err.Error())
	}
	hash := genHash(path + header.Filename)
	os.Rename(path+header.Filename, path+hash)
	log.Info("File uploaded successfully: " + header.Filename + "(" + hash + ")")
	return hash
}

func genHash(file string) string {
	f, err := os.Open(file)
	log.Check(log.FatalLevel, "Opening file"+file, err)
	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
