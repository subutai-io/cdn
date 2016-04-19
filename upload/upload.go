package upload

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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
  <form action="http://localhost:8080/` + repo + `/upload" method="post" enctype="multipart/form-data">
  <label for="file">Filename:</label>
  <input type="file" name="file" id="file">
  <input type="submit" name="submit" value="Submit">
  </form>
  </body>
  </html>
`
}

func Handler(w http.ResponseWriter, r *http.Request) string {
	hash := genHash()
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Warn(err.Error())
		return ""
	}
	defer file.Close()
	out, err := os.Create(path + hash)
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
	log.Info("File uploaded successfully: " + header.Filename)
	return hash
}

func genHash() string {
	hash := md5.New()
	hash.Write([]byte(time.Now().String()))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
