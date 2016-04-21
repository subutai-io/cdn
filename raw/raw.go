package raw

import (
	"net/http"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/download"
	"github.com/optdyn/gorjun/upload"
)

var (
	path = "/tmp/"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("raw")))
	} else if r.Method == "POST" {
		hash := upload.Handler(w, r)
		_, header, _ := r.FormFile("file")
		w.Write([]byte("Name: " + header.Filename + "\n"))
		db.Write("", hash, header.Filename)
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
