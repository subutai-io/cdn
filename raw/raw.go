package raw

import (
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/subutai-io/agent/log"

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		md5, sha256, owner := upload.Handler(w, r)
		if len(md5) == 0 || len(sha256) == 0 {
			return
		}
		info := map[string]string{
			"md5":    md5,
			"sha256": sha256,
			"type":   "raw",
		}
		r.ParseMultipartForm(32 << 20)
		if len(r.MultipartForm.Value["version"]) != 0 {
			info["version"] = r.MultipartForm.Value["version"][0]
		}
		_, header, _ := r.FormFile("file")
		id := uuid.NewV4().String()
		db.Write(owner, id, header.Filename, info)
		if len(r.MultipartForm.Value["private"]) > 0 && r.MultipartForm.Value["private"][0] == "true" {
			log.Info("Sharing " + md5 + " with " + owner)
			db.ShareWith(id, owner, owner)
		}

		w.Write([]byte(id))
		log.Info(header.Filename + " saved to raw repo by " + owner)
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	uri := strings.Replace(r.RequestURI, "/kurjun/rest/file/", "/kurjun/rest/raw/", 1)
	uri = strings.Replace(uri, "/kurjun/rest/raw/get", "/kurjun/rest/raw/download", 1)

	args := strings.Split(strings.TrimPrefix(uri, "/kurjun/rest/raw/"), "/")
	if len(args) > 0 && strings.HasPrefix(args[0], "download") {
		download.Handler("raw", w, r)
		return
	}
	if len(args) > 1 {
		if list := db.UserFile(args[0], args[1]); len(list) > 0 {
			http.Redirect(w, r, "/kurjun/rest/raw/download?id="+list[0], 302)
		}
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		if len(upload.Delete(w, r)) != 0 {
			w.Write([]byte("Removed"))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}
}

func Info(w http.ResponseWriter, r *http.Request) {
	info := download.Info("raw", r)
	if len(info) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}
	w.Write(info)
}
