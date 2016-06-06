package download

import (
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
)

type AptItem struct {
	Architecture string   `json:"architecture,omitempty"`
	Description  string   `json:"description,omitempty"`
	Filename     string   `json:"filename,omitempty"`
	Md5Sum       string   `json:"md5Sum,omitempty"`
	Name         string   `json:"name,omitempty"`
	Version      string   `json:"version,omitempty"`
	Size         string   `json:"size"`
	Owner        []string `json:"owner,omitempty"`
}

type RawItem struct {
	Md5Sum      string   `json:"md5Sum,omitempty"`
	Name        string   `json:"name,omitempty"`
	Package     string   `json:"package,omitempty"`
	Version     string   `json:"version,omitempty"`
	Fingerprint string   `json:"fingerprint"`
	Size        int64    `json:"size"`
	ID          string   `json:"id"`
	Owner       []string `json:"owner,omitempty"`
}

type ListItem struct {
	Architecture     string   `json:"architecture"`
	ConfigContents   string   `json:"configContents"`
	ID               string   `json:"id"`
	Md5Sum           string   `json:"md5Sum"`
	Name             string   `json:"name"`
	OwnerFprint      string   `json:"ownerFprint"`
	Package          string   `json:"package"`
	PackagesContents string   `json:"packagesContents"`
	Parent           string   `json:"parent"`
	Size             int64    `json:"size"`
	Version          string   `json:"version"`
	Owner            []string `json:"owner,omitempty"`
}

func Handler(repo string, w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	name := r.URL.Query().Get("name")
	if len(r.URL.Query().Get("id")) > 0 {
		hash = r.URL.Query().Get("id")
		if tmp := strings.Split(hash, "."); len(tmp) > 1 {
			hash = tmp[1]
		}
	}
	if len(hash) == 0 && len(name) == 0 {
		w.Write([]byte("Please specify hash or name"))
		return
	} else if len(name) != 0 {
		hash = db.LastHash(name, repo)
	}

	f, err := os.Open(config.Filepath + hash)
	defer f.Close()
	if log.Check(log.WarnLevel, "Opening file "+config.Filepath+hash, err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fi, _ := f.Stat()

	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && fi.ModTime().Unix() <= t.Unix() {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Last-Modified", fi.ModTime().Format(http.TimeFormat))
	w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))

	io.Copy(w, f)
}

func Info(repo string, r *http.Request) []byte {
	var item, js []byte

	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	rtype := r.URL.Query().Get("type")
	version := r.URL.Query().Get("version")

	list := db.Search(name)
	if len(id) > 0 {
		if len(strings.Split(id, ".")) > 1 {
			id = strings.Split(id, ".")[1]
		}
		list = map[string]string{id: ""}
	}

	counter := 0
	for k, _ := range list {
		info := db.Info(k)
		if info["type"] == repo {
			size, _ := strconv.ParseInt(info["size"], 10, 64)

			switch repo {
			case "template":
				item, _ = json.Marshal(ListItem{
					Name:         strings.Split(info["name"], "-")[0],
					ID:           info["owner"] + "." + k,
					OwnerFprint:  info["owner"],
					Parent:       info["parent"],
					Version:      info["version"],
					Architecture: strings.ToUpper(info["arch"]),
					Md5Sum:       k,
					Size:         size,
					Owner:        db.FileOwner(k),
				})
			case "apt":
				item, _ = json.Marshal(AptItem{
					Name:         info["name"],
					Md5Sum:       info["MD5sum"],
					Description:  info["Description"],
					Architecture: info["Architecture"],
					Version:      info["Version"],
					Size:         info["Size"],
					Owner:        db.FileOwner(k),
				})
			case "raw":
				item, _ = json.Marshal(RawItem{
					Name:    info["name"],
					ID:      k,
					Md5Sum:  k,
					Package: info["package"],
					Version: info["version"],
					Size:    size,
					Owner:   db.FileOwner(k),
				})
			}

			if name == strings.Split(info["name"], "-subutai-template")[0] || name == info["name"] {
				if (len(version) == 0 || info["version"] == version) && k == db.LastHash(info["name"], info["type"]) {
					if rtype == "text" {
						return ([]byte("public." + k))
					} else {
						return item
					}
				}
				continue
			}

			if counter++; counter > 1 {
				js = append(js, []byte(",")[0])
			}
			js = append(js, item...)
		}
	}
	if counter > 1 {
		js = append([]byte("["), js...)
		js = append(js, []byte("]")[0])
	}
	return js
}
