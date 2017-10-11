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
	"github.com/subutai-io/gorjun/upload"
)

// ListItem describes Gorjun entity. It can be APT package, Subutai template or Raw file.
type ListItem struct {
	ID           string            `json:"id"`
	Hash         hashsums          `json:"hash"`
	Size         int               `json:"size"`
	Name         string            `json:"name,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Owner        []string          `json:"owner,omitempty"`
	Parent       string            `json:"parent,omitempty"`
	Version      string            `json:"version,omitempty"`
	Filename     string            `json:"filename,omitempty"`
	Prefsize     string            `json:"prefsize,omitempty"`
	Signature    map[string]string `json:"signature,omitempty"`
	Description  string            `json:"description,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
}

type hashsums struct {
	Md5    string `json:"md5,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

// Handler provides download functionality for all artifacts.
func Handler(repo string, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	if len(id) == 0 && len(name) == 0 {
		io.WriteString(w, "Please specify id or name")
		return
	} else if len(name) != 0 {
		id = db.LastHash(name, repo)
	}

	if len(db.Read(id)) > 0 && !db.Public(id) && !db.CheckShare(id, db.CheckToken(r.URL.Query().Get("token"))) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}

	path := config.Storage.Path + id
	if md5, _ := db.Hash(id); len(md5) != 0 {
		path = config.Storage.Path + md5
	}

	f, err := os.Open(path)
	defer f.Close()

	if log.Check(log.WarnLevel, "Opening file "+config.Storage.Path+id, err) || len(id) == 0 {
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

	if name = db.Read(id); len(name) == 0 && len(config.CDN.Node) > 0 {
		httpclient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		resp, err := httpclient.Get(config.CDN.Node + "/kurjun/rest/template/info?id=" + id)
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
		w.Header().Set("Content-Disposition", "attachment; filename=\""+db.Read(id)+"\"")
	}

	io.Copy(w, f)
}

// Info returns JSON formatted list of elements. It allows to apply some filters to Search.
func Info(repo string, r *http.Request) []byte {
	var items []ListItem
	var info map[string]string
	p := []int{0, 1000}

	id := r.URL.Query().Get("id")
	tag := r.URL.Query().Get("tag")
	name := r.URL.Query().Get("name")
	page := r.URL.Query().Get("page")
	owner := r.URL.Query().Get("owner")
	token := r.URL.Query().Get("token")
	subname := r.URL.Query().Get("subname")
	version := r.URL.Query().Get("version")
	verified := r.URL.Query().Get("verified")
	if len(subname) != 0 {
		name = subname
	}

	list := db.Search(name)
	if len(tag) > 0 {
		listByTag, err := db.Tag(tag)
		log.Check(log.DebugLevel, "Looking for artifacts with tag "+tag, err)
		list = intersect(list, listByTag)
	}

	if len(id) > 0 {
		list = append(list[:0], id)
	} else if verified == "true" {
		items = append(items, getVerified(list, name, repo))
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

		if p[0]--; p[0] > 0 {
			continue
		}

		info = db.Info(k)

		if len(info["sha256"]) == 0 {
			if len(info["md5"]) == 0 {
				info["md5"] = info["id"]
			}
			info["sha256"] = upload.Hash(config.Storage.Path+info["md5"], "sha256")
			db.Write(db.FileField(info["id"], "owner")[0], info["id"], info["name"], map[string]string{"sha256": info["sha256"]})
		}
		item := formatItem(info, repo, name)

		if len(subname) == 0 && name == item.Name {
			if strings.HasSuffix(item.Version, version) || len(version) == 0 {
				items = []ListItem{item}
			}
		} else if len(version) == 0 || item.Version == version {
			items = append(items, item)
		}

		if len(items) >= p[1] {
			break
		}
	}
	if len(items) == 1 {
		items[0].Signature = db.FileSignatures(items[0].ID)
	}
	output, err := json.Marshal(items)
	if err != nil || string(output) == "null" {
		return nil
	}
	return output
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
				for _, owner := range db.FileField(info["id"], "owner") {
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
		Hash:         hashsums{Md5: info["md5"], Sha256: info["sha256"]},
		Name:         strings.Split(info["name"], "-subutai-template")[0],
		Tags:         db.FileField(info["id"], "tags"),
		Owner:        db.FileField(info["id"], "owner"),
		Version:      info["version"],
		Filename:     info["name"],
		Parent:       info["parent"],
		Prefsize:     info["prefsize"],
		Architecture: strings.ToUpper(info["arch"]),
		Description:  info["Description"],
	}
	item.Size, _ = strconv.Atoi(info["size"])

	if repo == "apt" {
		item.Version = info["Version"]
		item.Architecture = info["Architecture"]
	}
	if len(item.Hash.Md5) == 0 {
		item.Hash.Md5 = item.ID
	}
	return item
}

func intersect(listA, listB []string) (list []string) {
	mapA := map[string]bool{}
	for _, item := range listA {
		mapA[item] = true
	}
	for _, item := range listB {
		if mapA[item] {
			list = append(list, item)
		}
	}
	return list
}
