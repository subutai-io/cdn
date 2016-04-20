package download

import (
	"io"
	"net/http"
	"os"

	"github.com/optdyn/gorjun/db"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if len(hash) != 0 {
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
		w.Write([]byte("<p><a href=\"http://localhost:8080/" + repo + "/download?hash=" + k + "\">" + v + "</a></p>"))
	}
	w.Write([]byte("</body></html>"))
}

func Search(repo string, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	w.Write([]byte("<html><body>"))
	for k, v := range db.Search(query) {
		w.Write([]byte("<p><a href=\"http://localhost:8080/" + repo + "/download?hash=" + k + "\">" + v + "</a></p>"))
	}
	w.Write([]byte("</body></html>"))
}
