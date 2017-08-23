package template

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/subutai-io/agent/log"

	"fmt"

	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
)

func readTempl(hash string) (configfile string, err error) {
	var file bytes.Buffer
	f, err := os.Open(config.Storage.Path + hash)
	log.Check(log.WarnLevel, "Opening file "+config.Storage.Path+hash, err)
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}

	tr := tar.NewReader(gzf)

	for hdr, err := tr.Next(); err != io.EOF; hdr, err = tr.Next() {
		if hdr.Name == "config" {
			if _, err := io.Copy(&file, tr); err != nil {
				return "", err
			}
			break
		}
	}
	configfile = file.String()
	return configfile, nil
}

func getConf(hash string, configfile string) (t *download.ListItem) {
	t = &download.ListItem{ID: hash}

	for _, v := range strings.Split(configfile, "\n") {
		if line := strings.Split(v, "="); len(line) > 1 {
			line[0] = strings.TrimSpace(line[0])
			line[1] = strings.TrimSpace(line[1])

			switch line[0] {
			case "lxc.arch":
				t.Architecture = line[1]
			case "lxc.utsname":
				t.Name = line[1]
			case "subutai.parent":
				t.Parent = line[1]
			case "subutai.template.version":
				t.Version = line[1]
			case "subutai.template.size":
				t.Prefsize = line[1]
			case "subutai.template.description":
				t.Description = line[1]
			case "subutai.tags":
				t.Tags = []string{line[1]}
			}
		}
	}
	return
}

func Upload(w http.ResponseWriter, r *http.Request) {
	var hash, owner string
	if r.Method == "POST" {
		if hash, owner = upload.Handler(w, r); len(hash) == 0 {
			return
		}
		configfile, err := readTempl(hash)
		if err != nil || len(configfile) == 0 {
			log.Warn("Unable to read template config")
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("Unable to read configuration file. Is it a template archive?"))
			if db.Delete(owner, "template", hash) < 1 {
				f, _ := os.Stat(config.Storage.Path + hash)
				db.QuotaUsageSet(owner, -int(f.Size()))
				os.Remove(config.Storage.Path + hash)
			}
			return
		}
		t := getConf(hash, configfile)
		db.Write(owner, t.ID, t.Name+"-subutai-template_"+t.Version+"_"+t.Architecture+".tar.gz", map[string]string{
			"type":     "template",
			"arch":     t.Architecture,
			"tags":     strings.Join(t.Tags, ","),
			"parent":   t.Parent,
			"version":  t.Version,
			"prefsize": t.Prefsize,
		})
		w.Write([]byte(t.ID))
		log.Info(t.Name + " saved to template repo by " + owner)
		if len(r.MultipartForm.Value["private"]) > 0 && r.MultipartForm.Value["private"][0] == "true" {
			log.Info("Sharing " + hash + " with " + owner)
			db.ShareWith(hash, owner, owner)
		}
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	uri := strings.Replace(r.RequestURI, "/kurjun/rest/template/get", "/kurjun/rest/template/download", 1)
	args := strings.Split(strings.TrimPrefix(uri, "/kurjun/rest/template/"), "/")
	if len(args) > 0 && strings.HasPrefix(args[0], "download") {
		download.Handler("template", w, r)
		return
	}
	if len(args) > 1 {
		if list := db.UserFile(args[0], args[1]); len(list) > 0 {
			http.Redirect(w, r, "/kurjun/rest/template/download?id="+list[0], 302)
		}
	}
}

// func Torrent(w http.ResponseWriter, r *http.Request) {
// 	id := r.URL.Query().Get("id")
// 	if len(db.Read(id)) > 0 && !db.Public(id) && !db.CheckShare(id, db.CheckToken(r.URL.Query().Get("token"))) {
// 		w.WriteHeader(http.StatusNotFound)
// 		w.Write([]byte("Not found"))
// 		return
// 	}

// 	reader := torrent.Load([]byte(id))
// 	if reader == nil {
// 		return
// 	}
// 	mi, err := metainfo.Load(reader)
// 	if log.Check(log.WarnLevel, "Creating torrent for", err) {
// 		w.WriteHeader(http.StatusNotFound)
// 		w.Write([]byte("File not found"))
// 		return
// 	}

// 	err = mi.Write(w)
// 	log.Check(log.WarnLevel, "Writing to HTTP output", err)
// }

func Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		return
	}
	if info := download.Info("template", r); len(info) > 2 {
		w.Write(info)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		if len(upload.Delete(w, r)) != 0 {
			w.Write([]byte("Removed"))
		}
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Incorrect method"))
}

// Tag sets or removes additional tags for template artifact.
// It receives HTTP POST request for adding tags, and HTTP DELETE request for removing tags.
func Tag(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.ParseMultipartForm(32<<20) != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if code, err := addTag(r.MultipartForm.Value); err != nil {
			w.WriteHeader(code)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				log.Warn("Failed to write HTTP response")
			}
		}
	} else if r.Method == http.MethodDelete {
		if r.ParseMultipartForm(32<<20) != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if code, err := delTag(r.MultipartForm.Value); err != nil {
			w.WriteHeader(code)
			if _, err = w.Write([]byte(err.Error())); err != nil {
				log.Warn("Failed to write HTTP response")
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Incorrect method")); err != nil {
			log.Warn("Failed to write HTTP response")
		}

	}
}

func addTag(values map[string][]string) (int, error) {
	if len(values["token"]) > 0 {
		if user := db.CheckToken(values["token"][0]); len(values["token"][0]) == 0 || len(user) == 0 {
			return http.StatusUnauthorized, fmt.Errorf("Failed to authorize using provided token")
		} else if len(values["id"]) > 0 && len(values["tags"]) > 0 {
			if db.CheckRepo(user, "template", values["id"][0]) > 0 {
				db.Write(user, values["id"][0], "", map[string]string{"tags": values["tags"][0]})
				return http.StatusOK, nil
			}
		}
	}
	return http.StatusBadRequest, fmt.Errorf("Bad request")
}

func delTag(values map[string][]string) (int, error) {
	if len(values["token"]) > 0 {
		if user := db.CheckToken(values["token"][0]); len(values["token"][0]) == 0 || len(user) == 0 {
			return http.StatusUnauthorized, fmt.Errorf("Failed to authorize using provided token")
		} else if len(values["id"]) > 0 && len(values["tags"]) > 0 {
			if db.CheckRepo(user, "template", values["id"][0]) > 0 {
				db.RemoveTags(values["id"][0], values["tags"][0])
				return http.StatusOK, nil
			}
		}
	}
	return http.StatusBadRequest, fmt.Errorf("Bad request")
}
