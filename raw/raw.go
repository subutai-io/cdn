package raw

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		hash, owner, signature := upload.Handler(w, r)
		info := map[string]string{
			"type": "raw",
		}
		r.ParseMultipartForm(32 << 20)
		if len(r.MultipartForm.Value["version"]) != 0 {
			info["version"] = r.MultipartForm.Value["version"][0]
		}
		// info["signature"] = signature
		_, header, _ := r.FormFile("file")
		db.Write(owner, hash, header.Filename, info)
		w.Write([]byte(hash + "\n" + signature))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	//raw-files download handler will be here
	download.Handler("raw", w, r)
}

func List(w http.ResponseWriter, r *http.Request) {
	list := []download.RawItem{}
	for hash, _ := range db.List() {
		if info := db.Info(hash); info["type"] == "raw" {
			item := download.RawItem{
				ID:          "raw." + hash,
				Name:        info["name"],
				Fingerprint: info["owner"],
				Md5Sum:      hash,
				Owner:       db.FileOwner(hash),
			}
			item.Size, _ = strconv.ParseInt(info["size"], 10, 64)
			if version, exists := info["version"]; exists {
				item.Version = version
			}
			list = append(list, item)
		}
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
