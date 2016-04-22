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

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/download"
	"github.com/optdyn/gorjun/upload"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

type ListItem struct {
	Architecture   string `json:"architecture"`
	ConfigContents string `json:"configContents"`
	Extra          struct {
		Lxc_idMap           string `json:"lxc.id_map"`
		Lxc_include         string `json:"lxc.include"`
		Lxc_mount           string `json:"lxc.mount"`
		Lxc_mount_entry     string `json:"lxc.mount.entry"`
		Lxc_network_flags   string `json:"lxc.network.flags"`
		Lxc_network_hwaddr  string `json:"lxc.network.hwaddr"`
		Lxc_network_link    string `json:"lxc.network.link"`
		Lxc_network_type    string `json:"lxc.network.type"`
		Lxc_rootfs          string `json:"lxc.rootfs"`
		Subutai_config_path string `json:"subutai.config.path"`
		Subutai_git_branch  string `json:"subutai.git.branch"`
	} `json:"extra"`
	ID               string `json:"id"`
	Md5Sum           string `json:"md5Sum"`
	Name             string `json:"name"`
	OwnerFprint      string `json:"ownerFprint"`
	Package          string `json:"package"`
	PackagesContents string `json:"packagesContents"`
	Parent           string `json:"parent"`
	Size             int64  `json:"size"`
	Version          string `json:"version"`
}

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
	var hash string
	var config bytes.Buffer
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("template")))
	} else if r.Method == "POST" {
		if hash = upload.Handler(w, r); len(hash) == 0 {
			return
		}
		if config = readTempl(hash); len(config.String()) == 0 {
			return
		}
		t := getConf(hash, config)
		w.Write([]byte("Name: " + t.name + ", version: " + t.version + ", hash: " + t.hash + "\n"))
		db.Write("", t.hash, t.name+"-subutai-template_"+t.version+"_"+t.arch+".tar.gz",
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
	log.Info("md5")
	hash := md5.New()
	hash.Write([]byte(time.Now().String()))
	// w.Write([]byte("c9684cacea51e32d9304f5290b7e1b5e"))
	w.Write([]byte(fmt.Sprintf("%x", hash.Sum(nil))))
}

func List(w http.ResponseWriter, r *http.Request) {
	log.Info("list")
	list := make([]ListItem, 0)
	for hash, _ := range db.List() {
		var item ListItem
		for k, v := range db.Info(hash) {
			switch k {
			case "name":
				name := strings.Split(v, "-")
				if len(name) > 0 {
					item.Name = name[0]
				}
			case "arch":
				item.Architecture = v
			case "version":
				item.Version = v
			case "owner":
				item.OwnerFprint = v
			case "parent":
				item.Parent = v
			}
		}
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
