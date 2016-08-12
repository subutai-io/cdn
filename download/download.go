package download

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
)

type listItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AptItem struct {
	ID           string            `json:"id"`
	Size         string            `json:"size"`
	Name         string            `json:"name,omitempty"`
	Owner        []string          `json:"owner,omitempty"`
	Version      string            `json:"version,omitempty"`
	Filename     string            `json:"filename,omitempty"`
	Signature    map[string]string `json:"signature,omitempty"`
	Description  string            `json:"description,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
}

type RawItem struct {
	ID        string            `json:"id"`
	Size      int64             `json:"size"`
	Name      string            `json:"name,omitempty"`
	Owner     []string          `json:"owner,omitempty"`
	Version   string            `json:"version,omitempty"`
	Signature map[string]string `json:"signature,omitempty"`
}

type ListItem struct {
	ID           string            `json:"id"`
	Size         int64             `json:"size"`
	Name         string            `json:"name"`
	Filename     string            `json:"filename"`
	Parent       string            `json:"parent"`
	Version      string            `json:"version"`
	Owner        []string          `json:"owner,omitempty"`
	Architecture string            `json:"architecture"`
	Signature    map[string]string `json:"signature,omitempty"`
}

func Handler(repo string, w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	name := r.URL.Query().Get("name")
	if len(r.URL.Query().Get("id")) > 0 {
		hash = r.URL.Query().Get("id")
		if tmp := strings.Split(hash, "."); len(tmp) > 1 {
			hash = tmp[1]
		}
	}
	if len(hash) == 0 && len(name) == 0 {
		io.WriteString(w, "Please specify hash or name")
		return
	} else if len(name) != 0 {
		hash = db.LastHash(name, repo)
	}

	if !db.Public(hash) && !db.CheckShare(hash, db.CheckToken(r.URL.Query().Get("token"))) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}

	f, err := os.Open(config.Storage.Path + hash)
	defer f.Close()

	if log.Check(log.WarnLevel, "Opening file "+config.Storage.Path+hash, err) || len(hash) == 0 {
		if len(config.CDN.Node) > 0 {
			client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			resp, err := client.Get(config.CDN.Node + r.URL.RequestURI())
			if !log.Check(log.WarnLevel, "Getting file from CDN", err) {
				w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
				w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
				w.Header().Set("Last-Modified", resp.Header.Get("Last-Modified"))
				w.Header().Set("Content-Disposition", resp.Header.Get("Content-Disposition"))

				io.Copy(w, resp.Body)
				resp.Body.Close()
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "File not found")
		return
	}
	fi, _ := f.Stat()

	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && fi.ModTime().Unix() <= t.Unix() {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Last-Modified", fi.ModTime().Format(http.TimeFormat))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+db.Read(hash)+"\"")

	io.Copy(w, f)
}

func Info(repo string, r *http.Request) []byte {
	var item, js []byte

	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	rtype := r.URL.Query().Get("type")
	version := r.URL.Query().Get("version")

	list := db.Search(name)
	if len(id) > 0 {
		list = map[string]string{id: ""}
	}

	counter := 0
	for k, _ := range list {
		if !db.Public(k) && !db.CheckShare(k, db.CheckToken(r.URL.Query().Get("token"))) {
			// log.Warn("File " + k + " is not shared with " + db.CheckToken(r.URL.Query().Get("token")))
			continue
		}
		info := db.Info(k)
		if info["type"] == repo {
			size, _ := strconv.ParseInt(info["size"], 10, 64)

			switch repo {
			case "template":
				item, _ = json.Marshal(ListItem{
					ID:           k,
					Size:         size,
					Name:         strings.Split(info["name"], "-subutai-template")[0],
					Filename:     info["name"],
					Parent:       info["parent"],
					Version:      info["version"],
					Architecture: strings.ToUpper(info["arch"]),
					// Owner:        db.FileSignatures(k),
					Owner:     db.FileOwner(k),
					Signature: db.FileSignatures(k, name),
				})
			case "apt":
				item, _ = json.Marshal(AptItem{
					ID:           info["MD5sum"],
					Name:         info["name"],
					Description:  info["Description"],
					Architecture: info["Architecture"],
					Version:      info["Version"],
					Size:         info["Size"],
					// Owner:        db.FileSignatures(k),
					Owner:     db.FileOwner(k),
					Signature: db.FileSignatures(k, name),
				})
			case "raw":
				item, _ = json.Marshal(RawItem{
					ID:      k,
					Size:    size,
					Name:    info["name"],
					Version: info["version"],
					// Owner:   db.FileSignatures(k),
					Owner:     db.FileOwner(k),
					Signature: db.FileSignatures(k, name),
				})
			}

			if name == strings.Split(info["name"], "-subutai-template")[0] || name == info["name"] {
				if (len(version) == 0 || info["version"] == version) && k == db.LastHash(info["name"], info["type"]) {
					if rtype == "text" {
						log.Warn("Deprecated call to \"type=text\" endpoint")
						return ([]byte("public." + k))
					} else {
						return item
					}

				}
				continue
			}

			if counter++; counter > 1 {
				js = append(js, []byte(",")...)
			}
			js = append(js, item...)
		}
	}
	if counter > 1 {
		js = append([]byte("["), js...)
		js = append(js, []byte("]")...)
	}
	return js
}

// ProxyList retrieves list of artifacts from main CDN nodes if no data found in local database
// It creates simple JSON list of artifacts to provide it to Subutai Social.
func ProxyList(t string) []byte {
	if len(config.CDN.Node) == 0 {
		return nil
	}
	list := make([]listItem, 0)

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(config.CDN.Node + "/kurjun/rest/" + t + "/list")
	defer resp.Body.Close()
	if log.Check(log.WarnLevel, "Getting list from CDN", err) {
		return nil
	}

	rsp, err := ioutil.ReadAll(resp.Body)
	if log.Check(log.WarnLevel, "Reading from CDN response", err) {
		return nil
	}

	if log.Check(log.WarnLevel, "Decrypting request", json.Unmarshal([]byte(rsp), &list)) {
		return nil
	}

	output, err := json.Marshal(list)
	if log.Check(log.WarnLevel, "Marshaling list", err) {
		return nil
	}
	return output
}

// ProxyInfo retrieves information from main CDN nodes if no data found in local database
// It creates simple info JSON to provide it to Subutai Social.
func ProxyInfo(uri string) []byte {
	if len(config.CDN.Node) == 0 {
		return nil
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(config.CDN.Node + uri)
	defer resp.Body.Close()
	if log.Check(log.WarnLevel, "Getting list of templates from CDN", err) {
		return nil
	}

	rsp, err := ioutil.ReadAll(resp.Body)
	if log.Check(log.WarnLevel, "Reading from CDN response", err) {
		return nil
	}
	return rsp
}
