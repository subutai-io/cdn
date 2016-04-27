package raw

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

type RawItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Md5Sum      string `json:"md5Sum"`
	Version     string `json:"version"`
	Fingerprint string `json:"fingerprint"`
}

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

func List(w http.ResponseWriter, r *http.Request) {
	list := make([]RawItem, 0)
	for hash, _ := range db.List() {
		var item RawItem
		info := db.Info(hash)
		if info["type"] == "raw" {
			f, err := os.Open(config.Filepath + hash)
			if !log.Check(log.WarnLevel, "Opening file "+config.Filepath+hash, err) {
				fi, _ := f.Stat()
				f.Close()
				item.Size = fi.Size()
			}

			item.Fingerprint = info["owner"]
			item.Name = info["name"]
			item.Md5Sum = hash
			item.ID = "raw." + item.Md5Sum
			list = append(list, item)
		}
	}
	js, _ := json.Marshal(list)
	w.Write(js)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		log.Warn("Incorrect method")
		return
	}
	if len(upload.Delete(w, r)) != 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Removed"))
	}
}

func Info(w http.ResponseWriter, r *http.Request) {
	info := download.Info("raw", r)
	if len(info) != 0 {
		w.Write(info)
	} else {
		w.Write([]byte("Not found"))
	}
}
