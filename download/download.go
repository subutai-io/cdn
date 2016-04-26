package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/db"
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

func Handler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	name := r.URL.Query().Get("name")
	id := r.URL.Query().Get("id")
	if len(id) > 0 {
		hash = id
		tmp := strings.Split(id, ".")
		if len(tmp) == 2 {
			hash = tmp[1]
		}
	}
	if len(hash) == 0 && len(name) == 0 {
		w.Write([]byte("Please specify hash or name"))
		return
	} else if len(name) != 0 {
		hash = db.LastHash(name)
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	f, err := os.Open(path + hash)
	log.Check(log.WarnLevel, "Opening file "+path+hash, err)
	fi, _ := f.Stat()
	w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))

	defer f.Close()
	io.Copy(w, f)
}

func List(repo string, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html><body>"))
	for k, v := range db.List() {
		if db.Info(k)["type"] == repo {
			w.Write([]byte("<p><a href=\"/kurjun/rest/" + repo + "/download?hash=" + k + "\">" + v + "</a></p>"))
		}
	}
	w.Write([]byte("</body></html>"))
}

func Search(repo string, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	w.Write([]byte("<html><body>"))
	for k, v := range db.Search(query) {
		w.Write([]byte("<p><a href=\"/" + repo + "/download?hash=" + k + "\">" + v + "</a></p>"))
	}
	w.Write([]byte("</body></html>"))
}

func Info(repo string, w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	rtype := r.URL.Query().Get("type")
	version := r.URL.Query().Get("version")

	id := r.URL.Query().Get("id")
	ids := strings.Split(id, ".")
	if len(ids) > 1 {
		name = db.Read(ids[1])
	}

	if len(name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please specify template name"))
		return
	}

	for k, _ := range db.Search(name) {
		info := db.Info(k)
		if rtype == "text" {
			if strings.HasPrefix(info["name"], name+"-subutai-template") && (len(version) == 0 || info["version"] == version) {
				w.Write([]byte(info["owner"] + "." + k))
				return
			}
		} else {
			f, err := os.Open(path + k)
			log.Check(log.WarnLevel, "Opening file "+path+k, err)
			fi, _ := f.Stat()

			js, _ := json.Marshal(ListItem{
				Name:         strings.Split(info["name"], "-")[0],
				ID:           info["owner"] + "." + k,
				OwnerFprint:  info["owner"],
				Parent:       info["parent"],
				Version:      info["version"],
				Architecture: strings.ToUpper(info["arch"]),
				Size:         fi.Size(),
				Md5Sum:       k,
			})
			w.Write(js)
			return
		}
	}
	w.Write([]byte("Template not found"))
	w.WriteHeader(http.StatusNotFound)
}
