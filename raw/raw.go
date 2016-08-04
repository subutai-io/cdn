package raw

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		hash, owner := upload.Handler(w, r)
		info := map[string]string{
			"type": "raw",
			// "signature": signature,
		}
		r.ParseMultipartForm(32 << 20)
		if len(r.MultipartForm.Value["version"]) != 0 {
			info["version"] = r.MultipartForm.Value["version"][0]
		}
		_, header, _ := r.FormFile("file")
		db.Write(owner, hash, header.Filename, info)
		w.Write([]byte(hash))
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

func List(w http.ResponseWriter, r *http.Request) {
	list := []download.RawItem{}
	for hash, _ := range db.List() {
		if info := db.Info(hash); info["type"] == "raw" {
			item := download.RawItem{
				ID:   hash,
				Name: info["name"],
				// Owner:       db.FileSignatures(hash),
				Owner: db.FileOwner(hash),
			}
			item.Size, _ = strconv.ParseInt(info["size"], 10, 64)
			if version, exists := info["version"]; exists {
				item.Version = version
			}
			list = append(list, item)
		}
	}
	if len(list) == 0 {
		if js := download.ProxyList("raw"); js != nil {
			w.Write(js)
		}
		return
	}
	js, _ := json.Marshal(list)
	w.Write(js)

}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" && len(upload.Delete(w, r)) != 0 {
		w.Write([]byte("Removed"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
}

func Info(w http.ResponseWriter, r *http.Request) {
	info := download.Info("raw", r)
	if len(info) == 0 {
		w.Write([]byte("Not found"))
	}
	w.Write(info)
}
