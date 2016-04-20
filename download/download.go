package download

import (
	"io"
	"net/http"
	"os"

	"github.com/optdyn/gorjun/db"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/deb/"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	name := r.URL.Query().Get("name")
	if len(hash) != 0 {
		w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		f, err := os.Open(path + hash)
		log.Check(log.FatalLevel, "Opening file "+path+hash, err)
		defer f.Close()
		io.Copy(w, f)
	} else if len(name) != 0 {
		hash = db.LastHash(name)
		w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		f, err := os.Open(path + hash)
		log.Check(log.FatalLevel, "Opening file "+path+hash, err)
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
