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

	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
)

// ListItem describes Gorjun entity. It can be APT package, Subutai template or Raw file.
type ListItem struct {
	ID           string            `json:"id"`
	Size         int               `json:"size,omitempty"`
	Name         string            `json:"name,omitempty"`
	Owner        []string          `json:"owner,omitempty"`
	Parent       string            `json:"parent,omitempty"`
	Version      string            `json:"version,omitempty"`
	Filename     string            `json:"filename,omitempty"`
	Prefsize     string            `json:"prefsize,omitempty"`
	Signature    map[string]string `json:"signature,omitempty"`
	Description  string            `json:"description,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
}

// Handler provides download functionality for all artifacts.
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

	if len(db.Read(hash)) > 0 && !db.Public(hash) && !db.CheckShare(hash, db.CheckToken(r.URL.Query().Get("token"))) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}

	// if len(db.Read(hash)) == 0 && repo == "template" && !torrent.IsDownloaded(hash) {
	// 	torrent.AddTorrent(hash)
	// 	w.WriteHeader(http.StatusAccepted)
	// 	w.Write([]byte(torrent.Info(hash)))
	// 	return
	// }

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

	if name = db.Read(hash); len(name) == 0 && len(config.CDN.Node) > 0 {
		httpclient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		resp, err := httpclient.Get(config.CDN.Node + "/kurjun/rest/template/info?id=" + hash)
		if !log.Check(log.WarnLevel, "Getting info from CDN", err) {
			var info ListItem
			rsp, err := ioutil.ReadAll(resp.Body)
			if log.Check(log.WarnLevel, "Reading from CDN response", err) {
				w.WriteHeader(http.StatusNotFound)
				io.WriteString(w, "File not found")
				return
			}
			if !log.Check(log.WarnLevel, "Decrypting request", json.Unmarshal([]byte(rsp), &info)) {
				w.Header().Set("Content-Disposition", "attachment; filename=\""+info.Filename+"\"")
			}
			resp.Body.Close()
		}
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+db.Read(hash)+"\"")
	}

	io.Copy(w, f)
}

// Info returns JSON formatted list of elements. It allows to apply some filters to Search.
func Info(repo string, r *http.Request) []byte {
	var items []ListItem
	var info map[string]string
	p := []int{0, 1000}

	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	page := r.URL.Query().Get("page")
	owner := r.URL.Query().Get("owner")
	token := r.URL.Query().Get("token")
	version := r.URL.Query().Get("version")
	verified := r.URL.Query().Get("verified")

	list := db.Search(name)
	if len(id) > 0 {
		list = append(list[:0], id)
	} else if verified == "true" {
		items := append(items, getVerified(list, name, repo))
		output, err := json.Marshal(items)
		if err != nil || string(output) == "null" {
			return nil
		}
		return output
	}

	pstr := strings.Split(page, ",")
	p[0], _ = strconv.Atoi(pstr[0])
	if len(pstr) == 2 {
		p[1], _ = strconv.Atoi(pstr[1])
	}

	for _, k := range list {
		if (!db.Public(k) && !db.CheckShare(k, db.CheckToken(token))) ||
			(len(owner) > 0 && db.CheckRepo(owner, repo, k) == 0) ||
			db.CheckRepo("", repo, k) == 0 {
			continue
		}

		if p[0]--; p[0] >= 0 {
			continue
		}

		if name == "management" && repo == "template" {
			info = db.LatestTmpl(name, version)
			if len(info["name"]) == 0 {
				continue
			}
		} else {
			info = db.Info(k)
		}

		item := formatItem(info, repo, name)

		if strings.HasPrefix(info["name"], name+"-subutai-template") || name == info["name"] {
			if (len(version) == 0 || strings.Contains(info["version"], version)) && k == db.LastHash(info["name"], repo) {
				items = []ListItem{item}
				break
			}
			continue
		}

		if len(items) >= p[1] {
			break
		}
		items = append(items, item)
	}
	output, err := json.Marshal(items)
	if err != nil || string(output) == "null" {
		return nil
	}
	return output
}

// ProxyList retrieves list of artifacts from main CDN nodes if no data found in local database
// It creates simple JSON list of artifacts to provide it to Subutai Social.
func ProxyList(t string) []byte {
	if len(config.CDN.Node) == 0 {
		return nil
	}
	list := make([]ListItem, 0)

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(config.CDN.Node + "/kurjun/rest/" + t + "/list")
	if log.Check(log.WarnLevel, "Getting list from CDN", err) {
		return nil
	}
	defer resp.Body.Close()

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
	if log.Check(log.WarnLevel, "Getting list of templates from CDN", err) {
		return nil
	}
	defer resp.Body.Close()

	rsp, err := ioutil.ReadAll(resp.Body)
	if log.Check(log.WarnLevel, "Reading from CDN response", err) {
		return nil
	}
	return rsp
}

func in(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

func getVerified(list []string, name, repo string) ListItem {
	for _, k := range list {
		if info := db.Info(k); db.CheckRepo("", repo, k) > 0 {
			if info["name"] == name || (strings.HasPrefix(info["name"], name+"-subutai-template") && repo == "template") {
				for _, owner := range db.FileOwner(info["id"]) {
					if in(owner, []string{"subutai", "jenkins", "docker"}) {
						return formatItem(info, repo, name)
					}
				}
			}
		}
	}
	return ListItem{}
}

func formatItem(info map[string]string, repo, name string) ListItem {
	if len(info["prefsize"]) == 0 && repo == "template" {
		info["prefsize"] = "tiny"
	}

	item := ListItem{
		ID:           info["id"],
		Name:         strings.Split(info["name"], "-subutai-template")[0],
		Owner:        db.FileOwner(info["id"]),
		Version:      info["version"],
		Filename:     info["name"],
		Parent:       info["parent"],
		Prefsize:     info["prefsize"],
		Architecture: strings.ToUpper(info["arch"]),
		Signature:    db.FileSignatures(info["id"], name),
		Description:  info["Description"],
	}
	item.Size, _ = strconv.Atoi(info["size"])

	if repo == "apt" {
		item.Architecture = info["Architecture"]
		item.Version = info["Version"]
	}

	return item
}
