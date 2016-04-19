package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/optdyn/gorjun/db"
	"github.com/subutai-io/base/agent/log"
)

type template struct {
	name, parent, version, arch, hash string
}

var (
	path = "/tmp/"
)

var uploadPage string = `
  <html>
  <title>Go upload</title>
  <body>

  <form action="http://localhost:8080/template/upload" method="post" enctype="multipart/form-data">
  <label for="file">Filename:</label>
  <input type="file" name="file" id="file">
  <input type="submit" name="submit" value="Submit">
  </form>

  </body>
  </html>
`

func readTempl(hash string) (string, bytes.Buffer) {
	var config bytes.Buffer
	f, err := os.Open(path + hash)
	log.Check(log.FatalLevel, "Opening file "+path+hash, err)
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	log.Check(log.FatalLevel, "Creating gzip reader", err)

	tr := tar.NewReader(gzf)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		log.Check(log.FatalLevel, "Reading tar content", err)

		if hdr.Name == "config" {
			if _, err := io.Copy(&config, tr); err != nil {
				log.Fatal(err.Error())
			}
			break
		}
	}
	return hash, config
}

func getConf(hash string, config bytes.Buffer) (t *template) {
	t = &template{
		arch:    "lxc.arch",
		name:    "lxc.utsname",
		hash:    hash,
		parent:  "subutai.parent",
		version: "subutai.template.version",
	}

	for _, v := range strings.Split(config.String(), "\n") {
		line := strings.Split(v, "=")
		switch strings.Trim(line[0], " ") {
		case t.arch:
			t.arch = strings.Trim(line[1], " ")
		case t.name:
			t.name = strings.Trim(line[1], " ")
		case t.parent:
			t.parent = strings.Trim(line[1], " ")
		case t.version:
			t.version = strings.Trim(line[1], " ")
		}
	}
	return
}

func uploadHandler(w http.ResponseWriter, r *http.Request) string {
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
	log.Info("File uploaded successfully : " + header.Filename)
	return hash
}

func genHash() string {
	hash := md5.New()
	hash.Write([]byte(time.Now().String()))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func uploadTempl(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(uploadPage))
	} else if r.Method == "POST" {
		t := getConf(readTempl(uploadHandler(w, r)))
		w.Write([]byte("Name: " + t.name + ", version: " + t.version + ", hash: " + t.hash + "\n"))
		db.Write(t.hash, t.name+"-subutai-template_"+t.version+"_"+t.arch+".tar.gz")
		w.Write([]byte("Added to db: " + db.Read(t.hash)))
	}
}

func downloadTempl(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if len(hash) != 0 {
		w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		f, err := os.Open(path + hash)
		log.Check(log.FatalLevel, "Opening file "+path+hash, err)
		defer f.Close()
		io.Copy(w, f)
	} else {
		w.Write([]byte("Please specify name or hash"))
	}
}

func main() {
	defer db.Close()
	http.HandleFunc("/template/upload", uploadTempl)
	http.HandleFunc("/template/download", downloadTempl)
	http.ListenAndServe(":8080", nil)
}
