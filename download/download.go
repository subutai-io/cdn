package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/optdyn/gorjun/db"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

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
	if len(hash) != 0 {
		w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		f, err := os.Open(path + hash)
		log.Check(log.WarnLevel, "Opening file "+path+hash, err)
		fi, _ := f.Stat()
		w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))
		defer f.Close()
		io.Copy(w, f)
	} else if len(name) != 0 {
		hash = db.LastHash(name)
		w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		f, err := os.Open(path + hash)
		log.Check(log.WarnLevel, "Opening file "+path+hash, err)
		log.Check(log.WarnLevel, "Opening file "+path+hash, err)
		fi, _ := f.Stat()
		w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))
		defer f.Close()
		io.Copy(w, f)
	} else {
		w.Write([]byte("Please specify hash"))
	}
}

func List(repo string, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html><body>"))
	for k, v := range db.List() {
		w.Write([]byte("<p><a href=\"/" + repo + "/download?hash=" + k + "\">" + v + "</a></p>"))
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
	version := r.URL.Query().Get("version")

	if len(name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please specify template name"))
		return
	}

	for k, _ := range db.Search(name) {
		info := db.Info(k)
		if strings.HasPrefix(info["name"], name+"-subutai-template") {
			if len(version) == 0 {
				w.Write([]byte(info["owner"] + "." + k))
				return
			} else {
				if info["version"] == version {
					w.Write([]byte(info["owner"] + "." + k))
					return
				}
			}
		}
		w.Write([]byte("Template not found"))
		w.WriteHeader(http.StatusNotFound)
	}
}
