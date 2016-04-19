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
	t := getConf(readTempl(uploadHandler(w, r)))
	log.Info("Name: " + t.name + ", version: " + t.version + ", hash: " + t.hash)
	db.Write(t.hash, t.name)
	log.Info("Added to db: " + db.Read(t.hash))
}

func main() {
	defer db.Close()
	http.HandleFunc("/template/upload", uploadTempl)
	http.ListenAndServe(":8080", nil)
}
