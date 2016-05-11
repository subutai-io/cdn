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
	"strconv"
	"strings"
	"time"

	"github.com/subutai-io/base/agent/log"

	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

type Template struct {
	hash    string
	arch    string
	name    string
	parent  string
	version string
}

func readTempl(hash string) (configfile bytes.Buffer) {
	f, err := os.Open(config.Filepath + hash)
	log.Check(log.WarnLevel, "Opening file "+config.Filepath+hash, err)
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	log.Check(log.WarnLevel, "Creating gzip reader", err)

	tr := tar.NewReader(gzf)

	for hdr, err := tr.Next(); err != io.EOF; hdr, err = tr.Next() {
		if hdr.Name == "config" {
			if _, err := io.Copy(&configfile, tr); err != nil {
				log.Warn(err.Error())
			}
			break
		}
	}
	return configfile
}

func getConf(hash string, configfile bytes.Buffer) (t *Template) {
	t = &Template{hash: hash}

	for _, v := range strings.Split(configfile.String(), "\n") {
		if line := strings.Split(v, "="); len(line) > 1 {
			line[0] = strings.TrimSpace(line[0])
			line[1] = strings.TrimSpace(line[1])

			switch line[0] {
			case "lxc.arch":
				t.arch = line[1]
			case "lxc.utsname":
				t.name = line[1]
			case "subutai.parent":
				t.parent = line[1]
			case "subutai.template.version":
				t.version = line[1]
			}
		}
	}
	return
}

func Upload(w http.ResponseWriter, r *http.Request) {
	var hash, owner string
	var configfile bytes.Buffer
	if r.Method == "POST" {
		if hash, owner = upload.Handler(w, r); len(hash) == 0 {
			return
		}
		if configfile = readTempl(hash); len(configfile.String()) == 0 {
			return
		}
		t := getConf(hash, configfile)
		db.Write(owner, t.hash, t.name+"-subutai-template_"+t.version+"_"+t.arch+".tar.gz", map[string]string{
			"type":    "template",
			"arch":    t.arch,
			"parent":  t.parent,
			"version": t.version,
		})
		w.Write([]byte(t.hash))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	download.Handler(w, r)
}

func Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		return
	}

	if info := download.Info("template", r); len(info) != 0 {
		w.Write(info)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		if len(upload.Delete(w, r)) != 0 {
			w.Write([]byte("Removed"))
		}
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Incorrect method"))
}

func Md5(w http.ResponseWriter, r *http.Request) {
	hash := md5.New()
	hash.Write([]byte(time.Now().String()))
	w.Write([]byte(fmt.Sprintf("%x", hash.Sum(nil))))
}

func List(w http.ResponseWriter, r *http.Request) {
	list := make([]download.ListItem, 0)
	for hash, _ := range db.List() {
		if info := db.Info(hash); info["type"] == "template" {
			item := download.ListItem{
				ID:           "public." + hash,
				Name:         strings.Split(info["name"], "-")[0],
				Md5Sum:       hash,
				Parent:       info["parent"],
				Version:      info["version"],
				OwnerFprint:  info["owner"],
				Architecture: strings.ToUpper(info["arch"]),
				Owner:        db.FileOwner(hash),
			}
			item.Size, _ = strconv.ParseInt(info["size"], 10, 64)
			list = append(list, item)
		}
	}
	js, _ := json.Marshal(list)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Write(js)
}
