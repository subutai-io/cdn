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

	"bufio"
	"code.cloudfoundry.org/archiver/extractor"
	"errors"
	"github.com/jhoonb/archivex"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"
	"io/ioutil"
	"net/url"
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
	io.Copy(&file, f)
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

func getConfig(hash string, configfile, id string) (t *download.ListItem) {
	t = &download.ListItem{ID: id}
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
			"type":           "template",
			"arch":           t.Architecture,
			"md5":            md5,
			"sha256":         sha256,
			"tags":           strings.Join(t.Tags, ","),
			"parent":         t.Parent,
			"parent-version": t.ParentVersion,
			"parent-owner":   t.ParentOwner,
			"version":        t.Version,
			"prefsize":       t.Prefsize,
			"Description":    t.Description,
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

func ModifyConfig(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	name := r.URL.Query().Get("name")
	owner := strings.ToLower(db.CheckToken(token))
	if len(token) == 0 || len(owner) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting unauthorized owner request")
		return
	}
	if owner != "subutai" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only allowed users can update template config"))
		log.Warn(r.RemoteAddr + " - rejecting update request")
		return
	}
	list := db.Search(name)
	for _, k := range list {
		if db.CheckRepo("", "template", k) == 0 {
			continue
		}

		item := download.FormatItem(db.Info(k), "template", "")
		md5 := item.Hash.Md5
		configPath := config.Storage.Path + "/tmp/foo/config"

		err := decompress(config.Storage.Path+md5, config.Storage.Path+"/tmp/foo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can't decompress this template " + config.Storage.Path + md5))
			return
		}
		err = appendConfig(configPath, item)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can't find parent of this template , ID = " + item.ID + "\n" +
				"Name =" + item.Name + "\n" +
				"Parent =" + item.Parent + "\n"))
			return
		}
		err = compress(config.Storage.Path+"/tmp/foo", config.Storage.Path+"/tmp/foo.tar.gz")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can't compress this template " + config.Storage.Path + md5))
			return
		}
		err = updateMetaDB(item.ID, item.Owner[0], item.Hash.Md5, item.Filename, configPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can't update metadate of template " + config.Storage.Path + md5))
			return
		}
		err = os.RemoveAll(config.Storage.Path + "/tmp/foo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can't remove this  " + config.Storage.Path + "/tmp/foo" + "directory"))
			return
		}
		os.RemoveAll(config.Storage.Path + md5)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can't remove this  " + config.Storage.Path + md5 + "directory"))
			return
		}
	}
}

func appendConfig(confPath string, item download.ListItem) error {
	templateParent := item.Parent
	list := db.Search(templateParent)
	latestVerified := download.GetVerified(list, templateParent, "template", "")
	if latestVerified.ID == "" {
		return errors.New("Can't find parent of template")
	}
	SetContainerConf(confPath, [][]string{
		{"subutai.template", item.Name},
		{"subutai.template.owner", item.Owner[0]},
		{"subutai.parent", latestVerified.Name},
		{"subutai.parent.owner", latestVerified.Owner[0]},
		{"subutai.parent.version", latestVerified.Version},
	})
	return nil
}

func updateMetaDB(id, owner, hash, filename, configPath string) error {
	md5sum := upload.Hash(config.Storage.Path + "/tmp/foo.tar.gz")
	sha256sum := upload.Hash(config.Storage.Path+"/tmp/foo.tar.gz", "sha256")
	if len(md5sum) == 0 || len(sha256sum) == 0 {
		log.Warn("Failed to calculate hash for " + hash)
		return errors.New("Failed to calculate")
	}
	configfile, _ := readTemplate(configPath)
	t := getConfig(hash, configfile, id)
	filename = t.Name + "-subutai-template_" + t.Version + "_" + t.Architecture + ".tar.gz"
	t.Signature = db.FileSignatures(id)
	fmt.Println(t.ParentVersion)
	fmt.Println(t.ParentOwner)
	db.Edit(owner, id, filename, map[string]string{
		"type":           "template",
		"arch":           t.Architecture,
		"md5":            md5sum,
		"sha256":         sha256sum,
		"tags":           strings.Join(t.Tags, ","),
		"parent":         t.Parent,
		"parent-owner":   t.ParentOwner,
		"parent-version": t.ParentVersion,
		"version":        t.Version,
		"prefsize":       t.Prefsize,
		"Description":    t.Description,
		"signature":      t.Signature[owner],
	})

	err := os.Rename(config.Storage.Path+"/tmp/foo.tar.gz", config.Storage.Path+md5sum)

	if err != nil {
		fmt.Println(err)
		return errors.New("Can't rename tar file")
	}
	return nil
}

func decompress(file string, folder string) error {
	tgz := extractor.NewTgz()
	err := tgz.Extract(file, folder)
	if err != nil {
		return err
	}
	return nil
}

func compress(folder, file string) error {
	archive := new(archivex.TarFile)
	err := archive.Create(file)
	err = archive.AddAll(folder, false)
	if err != nil {
		return err
	}
	archive.Close()
	return nil
}

func SetContainerConf(confPath string, conf [][]string) error {

	newconf := ""

	file, err := os.Open(confPath)
	if log.Check(log.DebugLevel, "Opening container config "+confPath, err) {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(bufio.NewReader(file))
	for scanner.Scan() {
		newline := scanner.Text() + "\n"
		for i := 0; i < len(conf); i++ {
			line := strings.Split(scanner.Text(), "=")
			if len(line) > 1 && strings.Trim(line[0], " ") == conf[i][0] {
				if newline = ""; len(conf[i][1]) > 0 {
					newline = conf[i][0] + " = " + conf[i][1] + "\n"
				}
				conf = append(conf[:i], conf[i+1:]...)
				break
			}
		}
		newconf = newconf + newline
	}

	for i := range conf {
		if conf[i][1] != "" {
			newconf = newconf + conf[i][0] + " = " + conf[i][1] + "\n"
		}
	}
	return ioutil.WriteFile(confPath, []byte(newconf), 0644)
}
