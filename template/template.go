package template

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
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

func readTempl(hash string) bytes.Buffer {
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
	return config
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
	var hash, owner string
	var config bytes.Buffer
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("template")))
	} else if r.Method == "POST" {
		if hash, owner = upload.Handler(w, r); len(hash) == 0 {
			return
		}
		if config = readTempl(hash); len(config.String()) == 0 {
			return
		}
		t := getConf(hash, config)
		w.Write([]byte("Name: " + t.name + ", version: " + t.version + ", hash: " + t.hash + "\n"))
		db.Write(owner, t.hash, t.name+"-subutai-template_"+t.version+"_"+t.arch+".tar.gz",
			map[string]string{
				"arch":    t.arch,
				"version": t.version,
				"parent":  t.parent,
				"type":    "template",
			})
		w.Write([]byte("Added to db: " + db.Read(t.hash)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	download.Handler(w, r)
}

func Show(w http.ResponseWriter, r *http.Request) {
	download.List("template", w, r)
}

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		log.Warn("Incorrect method")
		return
	}
	download.Search("template", w, r)
}

func Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		log.Warn("Incorrect method")
		return
	}
	download.Info("template", w, r)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		log.Warn("Incorrect method")
		return
	}
	if len(upload.Delete(w, r)) != 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Removed"))
	}
}

func Md5(w http.ResponseWriter, r *http.Request) {
	hash := md5.New()
	hash.Write([]byte(time.Now().String()))
	// w.Write([]byte("c9684cacea51e32d9304f5290b7e1b5e"))
	w.Write([]byte(fmt.Sprintf("%x", hash.Sum(nil))))
}

func List(w http.ResponseWriter, r *http.Request) {
	list := make([]download.ListItem, 0)
	for hash, _ := range db.List() {
		var item download.ListItem
		info := db.Info(hash)
		if info["type"] != "template" {
			continue
		}

		name := strings.Split(info["name"], "-")
		if len(name) > 0 {
			item.Name = name[0]
		}
		item.Architecture = strings.ToUpper(info["arch"])
		item.Version = info["version"]
		item.OwnerFprint = info["owner"]
		item.Parent = info["parent"]

		f, err := os.Open(path + hash)
		log.Check(log.WarnLevel, "Opening file "+path+hash, err)
		fi, _ := f.Stat()
		item.Size = fi.Size()
		item.Md5Sum = hash
		item.ID = item.OwnerFprint + "." + item.Md5Sum
		list = append(list, item)
	}
	js, _ := json.Marshal(list)
	w.Write(js)
}
