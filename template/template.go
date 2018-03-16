package template

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/subutai-io/agent/log"

	"fmt"

	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
	"net/url"
	"os/exec"
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

func readTemplate(dir string) (configfile string, err error) {
	var file bytes.Buffer
	f, err := os.Open(dir)
	log.Check(log.WarnLevel, "Opening file "+dir, err)
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
	my_uuid, _ := uuid.NewV4()
	t = &download.ListItem{ID: my_uuid.String()}
	t.Hash.Md5 = hash
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

func getConfig(hash string, configfile string) (t *download.ListItem) {
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
			case "subutai.parent.owner":
				t.ParentOwner = line[1]
			case "subutai.parent.version":
				t.ParentVersion = line[1]
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
	if r.Method == "POST" {
		md5, sha256, owner := upload.Handler(w, r)
		if len(md5) == 0 || len(sha256) == 0 {
			return
		}
		configfile, err := readTempl(md5)
		if err != nil || len(configfile) == 0 {
			log.Warn("Unable to read template config")
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte("Unable to read configuration file. Is it a template archive?"))
			if db.Delete(owner, "template", md5) < 1 {
				f, _ := os.Stat(config.Storage.Path + md5)
				db.QuotaUsageSet(owner, -int(f.Size()))
				os.Remove(config.Storage.Path + md5)
			}
			return
		}
		t := getConf(md5, configfile)
		filename := t.Name + "-subutai-template_" + t.Version + "_" + t.Architecture + ".tar.gz"
		db.Write(owner, t.ID, filename, map[string]string{
			"type":        "template",
			"arch":        t.Architecture,
			"md5":         md5,
			"sha256":      sha256,
			"tags":        strings.Join(t.Tags, ","),
			"parent":      t.Parent,
			"version":     t.Version,
			"prefsize":    t.Prefsize,
			"Description": t.Description,
		})
		if len(r.MultipartForm.Value["private"]) > 0 && r.MultipartForm.Value["private"][0] == "true" {
			log.Info("Sharing " + t.ID + " with " + owner)
			db.ShareWith(t.ID, owner, owner)
		}

		w.Write([]byte(t.ID))
		log.Info(t.Name + " saved to template repo by " + owner)

		if IDs := db.UserFile(owner, filename); len(IDs) > 0 {
			for _, ID := range IDs {
				if ID == t.ID {
					continue
				}
				item := download.FormatItem(db.Info(ID), "template", filename)
				if db.Delete(owner, "template", item.ID) < 1 {
					f, _ := os.Stat(config.Storage.Path + item.Hash.Md5)
					db.QuotaUsageSet(owner, -int(f.Size()))
					if item.Hash.Md5 != t.Hash.Md5 {
						os.Remove(config.Storage.Path + item.Hash.Md5)
					}
				}
			}
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
		parsedUrl, _ := url.Parse(uri)
		parameters, _ := url.ParseQuery(parsedUrl.RawQuery)
		var token string
		if len(parameters["token"]) > 0 {
			token = parameters["token"][0]
		}
		owner := args[0]
		file := strings.Split(args[1], "?")[0]
		if list := db.UserFile(owner, file); len(list) > 0 {
			http.Redirect(w, r, "/kurjun/rest/template/download?id="+list[0]+"&token="+token, 302)
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
	ModifyConfig()
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

func ModifyConfig() {
	list := db.Search("")
	for _, k := range list {
		if db.CheckRepo("", "template", k) == 0 {
			continue
		}
		item := download.FormatItem(db.Info(k), "template", "")
		if !exists(config.Storage.Path + "myfolder") {
			cmd := exec.Command("bash", "-c", "mkdir myfolder "+item.Hash.Md5)
			cmd.Dir = config.Storage.Path
			cmd.Run()
		}
		cmd := exec.Command("bash", "-c", "tar -C "+config.Storage.Path+"/myfolder"+" -xvzf "+item.Hash.Md5)
		configfile, err := readTemplate(config.Storage.Path + "/myfolder/" + item.Hash.Md5)

		cmd.Dir = config.Storage.Path
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			log.Info("Can't run tar command")
		}
		appendFile(prepareMetaData(item))

		cmd = exec.Command("bash", "-c",
			"cd "+config.Storage.Path+"/myfolder;"+
				"tar -rvf "+item.Hash.Md5+" *")
		//cmd.Dir = config.Storage.Path
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			log.Info("Can't run tar command")
		}
		md5sum := upload.Hash(config.Storage.Path + item.Hash.Md5)
		cmd = exec.Command("bash", "-c",
			"mv "+item.Hash.Md5+" ..;"+
				"rm -rf *;"+
				"cd ..; mv "+item.Hash.Md5+" "+md5sum)
		cmd.Dir = config.Storage.Path
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			log.Info("Can't run tar command")
		}

		updateMetaDB(item.ID, item.Owner[0], item.ParentOwner, item.ParentVersion, item.Hash.Md5, item.Filename)
	}
}

func appendFile(metadata string) {
	file, err := os.OpenFile(config.Storage.Path+"/myfolder/"+"config", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("failed opening file: %s", err)
	}
	defer file.Close()

	_, err = file.WriteString(metadata)
	if err != nil {
		log.Fatal("failed writing to file: %s", err)
	}
}

func prepareMetaData(item download.ListItem) string {
	templateParent := item.Parent
	list := db.Search(templateParent)
	latestVerified := download.GetVerified(list, templateParent, "template", "")
	metadata := "subutai.template = " + item.Name + "\n" +
		"subutai.template.owner = " + item.Owner[0] + "\n" +
		"subutai.parent.owner = " + latestVerified.Name + "\n" +
		"subutai.parent.version = " + latestVerified.Version + "\n"
	return metadata
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func updateMetaDB(id, owner, parentowner, parentversion, hash, filename string) {
	md5sum := upload.Hash(config.Storage.Path + hash)
	sha256sum := upload.Hash(config.Storage.Path+hash, "sha256")
	if len(md5sum) == 0 || len(sha256sum) == 0 {
		log.Warn("Failed to calculate hash for " + hash)
		return
	}
	t := getConf(hash, configfile)
	filename := t.Name + "-subutai-template_" + t.Version + "_" + t.Architecture + ".tar.gz"
	db.Write(owner, t.ID, filename, map[string]string{
		"type":        "template",
		"arch":        t.Architecture,
		"md5":         md5,
		"sha256":      sha256,
		"tags":        strings.Join(t.Tags, ","),
		"parent":      t.Parent,
		"version":     t.Version,
		"prefsize":    t.Prefsize,
		"Description": t.Description,
	})
}
