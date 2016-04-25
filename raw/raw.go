package raw

import (
	"net/http"

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

var (
	path = "/tmp/"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("raw")))
	} else if r.Method == "POST" {
		hash, owner := upload.Handler(w, r)
		_, header, _ := r.FormFile("file")
		w.Write([]byte("Name: " + header.Filename + "\n"))
		db.Write(owner, hash, header.Filename, map[string]string{"type": "raw"})
		w.Write([]byte("Added to db: " + db.Read(hash)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	//raw-files download handler will be here
	download.Handler(w, r)
}

func Show(w http.ResponseWriter, r *http.Request) {
	//raw-files list handler will be here
	download.List("raw", w, r)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if len(upload.Delete(w, r)) != 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Removed"))
	}
}
