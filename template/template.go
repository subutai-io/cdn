package template

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/download"
	"github.com/optdyn/gorjun/upload"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

type Template struct {
	hash    string
	arch    string
	name    string
	parent  string
	version string
}

func readTempl(hash string) (string, bytes.Buffer) {
	var config bytes.Buffer
	f, err := os.Open(path + hash)
	log.Check(log.WarnLevel, "Opening file "+path+hash, err)
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	log.Check(log.WarnLevel, "Creating gzip reader", err)

	tr := tar.NewReader(gzf)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		log.Check(log.WarnLevel, "Reading tar content", err)

		if hdr.Name == "config" {
			if _, err := io.Copy(&config, tr); err != nil {
				log.Warn(err.Error())
			}
			break
		}
	}
	return hash, config
}

func getConf(hash string, config bytes.Buffer) (t *Template) {
	t = &Template{
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

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("template")))
	} else if r.Method == "POST" {
		t := getConf(readTempl(upload.Handler(w, r)))
		w.Write([]byte("Name: " + t.name + ", version: " + t.version + ", hash: " + t.hash + "\n"))
		db.Write("", t.hash, t.name+"-subutai-template_"+t.version+"_"+t.arch+".tar.gz",
			map[string]string{
				"arch":    t.arch,
				"version": t.version,
				"parent":  t.parent,
			})
		w.Write([]byte("Added to db: " + db.Read(t.hash)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	//templates download handler will be here
	download.Handler(w, r)
}

func Show(w http.ResponseWriter, r *http.Request) {
	//templates list handler will be here
	download.List("template", w, r)
}

func Search(w http.ResponseWriter, r *http.Request) {
	download.Search("template", w, r)
}
