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

	"github.com/blang/semver"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
)

// ListItem describes Gorjun entity. It can be APT package, Subutai template or Raw file.
type ListItem struct {
	ID            string            `json:"id"`
	Hash          hashsums          `json:"hash"`
	Size          int               `json:"size"`
	Name          string            `json:"name,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Owner         []string          `json:"owner,omitempty"`
	Parent        string            `json:"parent,omitempty"`
	ParentVersion string            `json:"parent-version,omitempty"`
	ParentOwner   string            `json:"parent-owner,omitempty"`
	Version       string            `json:"version,omitempty"`
	Filename      string            `json:"filename,omitempty"`
	Prefsize      string            `json:"prefsize,omitempty"`
	Signature     map[string]string `json:"signature,omitempty"`
	Description   string            `json:"description,omitempty"`
	Architecture  string            `json:"architecture,omitempty"`
	Date          time.Time         `json:"upload-date-formatted"`
	Timestamp     string            `json:"upload-date-timestamp,omitempty"`
}

type hashsums struct {
	Md5    string `json:"md5,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

// Handler provides download functionality for all artifacts.
func Handler(repo string, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	name := strings.ToLower(r.URL.Query().Get("name"))
	tag := r.URL.Query().Get("tag")

	tagSplit := strings.Split(tag, ",")
	if len(id) == 0 && len(name) == 0 {
		io.WriteString(w, "Please specify id or name")
		return
	}
	if len(name) != 0 {
		if len(tag) != 0 {
			if len(tagSplit) > 1 {
				listbyTag := db.UnionByTags(tagSplit, repo)
				for _, t := range listbyTag {
					if db.NameByHash(t) == name {
						id = t
					}
				}
			} else {
				listbyTag := db.SearchByOneTag(tag, repo)
				for _, t := range listbyTag {
					if db.NameByHash(t) == name {
						id = t
					}
				}
			}
		} else {
			id = db.LastHash(name, repo)
		}
	}

	if len(db.NameByHash(id)) > 0 && !db.IsPublic(id) && !db.CheckShare(id, db.TokenOwner(strings.ToLower(r.URL.Query().Get("token")))) {
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
	if name = db.NameByHash(id); len(name) == 0 && len(config.CDN.Node) > 0 {
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
		w.Header().Set("Content-Disposition", "attachment; filename=\""+db.NameByHash(id)+"\"")
	}
	io.Copy(w, f)
}

// Info returns JSON formatted list of elements. It allows to apply some filters to Search.
func Info(repo string, r *http.Request) []byte {
	log.Debug(fmt.Sprintf("Received info request.\n\nrepo: %+v\n\nr: %+v\n\n", repo, r))
	var items []ListItem
	var itemLatestVersion ListItem
	p := []int{0, 1000}
	id := r.URL.Query().Get("id")
	tag := r.URL.Query().Get("tag")
	name := strings.ToLower(r.URL.Query().Get("name"))
	page := r.URL.Query().Get("page")
	owner := strings.ToLower(r.URL.Query().Get("owner"))
	token := strings.ToLower(r.URL.Query().Get("token"))
	version := r.URL.Query().Get("version")
	verified := r.URL.Query().Get("verified")
	version = processVersion(version)
	list := make([]string, 0)
	if id != "" {
		//		log.Debug(fmt.Sprintf("id was provided"))
		name = db.NameByHash(id)
		list = append(list, id)
	} else {
		if name == "" {
			log.Warn(fmt.Sprintf("Both id and name were not provided"))
			return nil
		}
		list = db.SearchName(name)
		if owner != "" {
			log.Debug(fmt.Sprintf("If #1"))
			//			log.Debug(fmt.Sprintf("#1 List\n\n%+v\n\nintersected with\n\n%+v\n\nResulting in\n\n%+v\n\n", list, db.OwnerFilesByRepo(owner, repo), intersect(list, db.OwnerFilesByRepo(owner, repo))))
			list = intersect(list, db.OwnerFilesByRepo(owner, repo))
		}
		if token != "" && ((owner == "" && db.TokenOwner(token) != "") || (owner != "" && db.TokenOwner(token) == owner)) {
			log.Debug(fmt.Sprintf("If #2"))
			if owner == "" {
				//				log.Debug(fmt.Sprintf("#2.1 List\n\n%+v\n\nintersected with\n\n%+v\n\nResulting in\n\n%+v\n\n", list, db.TokenFilesByRepo(token, repo), intersect(list, db.TokenFilesByRepo(token, repo))))
				list = intersect(list, db.TokenFilesByRepo(token, repo))
				if len(list) == 0 {
					log.Debug(fmt.Sprintf("If #2.1.1"))
					list = db.SearchName(name)
					verified = "true"
				}
			} else if owner != "" && len(list) == 0 {
				//				log.Debug(fmt.Sprintf("#2.2 List\n\n%+v\n\nintersected with\n\n%+v\n\nResulting in\n\n%+v\n\n", list, db.TokenFilesByRepo(token, repo), intersect(list, db.TokenFilesByRepo(token, repo))))
				list = intersect(db.SearchName(name), db.TokenFilesByRepo(token, repo))
			}
		}
	}
	if owner == "" && (token == "" || (token != "" && db.TokenOwner(token) == "")) {
		//		log.Debug(fmt.Sprintf("If #3"))
		verified = "true"
	}
	list = unique(list)
	if tag != "" {
		log.Debug(fmt.Sprintf("Filtering with tag %+v", tag))
		listByTag, err := db.Tag(tag)
		log.Check(log.DebugLevel, "Looking for artifacts with tag "+tag, err)
		list = intersect(list, listByTag)
	}
	if verified == "true" {
		log.Debug(fmt.Sprintf("Filtering files to verified users. List: %+v", list))
		itemLatestVersion = GetVerified(list, name, repo, version)
		if itemLatestVersion.ID != "" {
			items = append(items, itemLatestVersion)
			items[0].Signature = db.FileSignatures(items[0].ID)
		}
		log.Debug(fmt.Sprintf("info collected. items (1): %+v", items))
		output, err := json.Marshal(items)
		if err == nil && len(items) > 0 && items[0].ID != "" {
			return output
		}
		return nil
	}
	pstr := strings.Split(page, ",")
	p[0], _ = strconv.Atoi(pstr[0])
	if len(pstr) == 2 {
		p[1], _ = strconv.Atoi(pstr[1])
	}
	latestVersion, _ := semver.Make("")
	log.Debug(fmt.Sprintf("list to be checked: %+v", list))
	for _, k := range list {
		if (!db.IsPublic(k) && !db.CheckShare(k, db.TokenOwner(token))) ||
			(db.IsPublic(k) && len(owner) > 0 && db.CheckRepo(owner, []string{repo}, k) == 0) ||
			db.CheckRepo("", []string{repo}, k) == 0 {
			log.Debug(fmt.Sprintf("File %+v (name: %+v, owner: %+v, token: %+v) is ignored: %+v || %+v || %+v", k, db.NameByHash(k), owner, db.TokenOwner(token), !db.IsPublic(k) && !db.CheckShare(k, db.TokenOwner(token)), db.IsPublic(k) && len(owner) > 0 && db.CheckRepo(owner, []string{repo}, k) == 0, db.CheckRepo("", []string{repo}, k) == 0))
			continue
		}
		if p[0]--; p[0] > 0 {
			continue
		}
		item := FormatItem(db.Info(k), repo)
		if (name == item.Name /* || (repo == "template" && strings.HasPrefix(item.Name, name+"-subutai-template") )*/) &&
			(version == "" || (version != "" && item.Version == version)) {
			items = []ListItem{item}
			itemVersion, _ := semver.Make(item.Version)
			if itemVersion.GTE(latestVersion) {
				latestVersion = itemVersion
				itemLatestVersion = item
			}
		}
		if len(items) >= p[1] {
			break
		}
	}
	if len(items) == 1 {
		if version == "" && repo == "template" && itemLatestVersion.ID != "" {
			log.Debug(fmt.Sprintf("version == \"\" && repo == \"template\" && itemLatestVersion.ID != \"\" returns true.\nitemLatestVersion.ID = %+v", itemLatestVersion))
			items[0] = itemLatestVersion
		}
		items[0].Signature = db.FileSignatures(items[0].ID)
	}
	log.Debug(fmt.Sprintf("info collected (repo: %+v, r: %+v). items (2):\n\n\n", repo, r))
	for i := 0; i < len(items); i++ {
		log.Debug(fmt.Sprintf("\nItem #%+v: %+v\n", i, items[i]))
	}
	output, err := json.Marshal(items)
	if err != nil || string(output) == "null" {
		return nil
	}
	return output
}

func List(repo string, r *http.Request) []byte {
	log.Debug(fmt.Sprintf("Received list request.\nrepo: %+v\nr: %+v", repo, r))
	var items []ListItem
	p := []int{0, 1000}
	tag := r.URL.Query().Get("tag")
	name := strings.ToLower(r.URL.Query().Get("name"))
	page := r.URL.Query().Get("page")
	owner := strings.ToLower(r.URL.Query().Get("owner"))
	token := strings.ToLower(r.URL.Query().Get("token"))
	version := r.URL.Query().Get("version")
	list := make([]string, 0)
	list = db.SearchName(name)
	if owner != "" {
		log.Debug(fmt.Sprintf("Owner not empty: %+v. Gathering his public files"), owner)
		list = db.OwnerFilesByRepo(owner, repo)
	}
	if token != "" && db.TokenOwner(token) != "" {
		log.Debug(fmt.Sprintf("%+v token's owner not empty: %+v. Adding his private/shared files", token, db.TokenOwner(token)))
		list = append(list, db.TokenFilesByRepo(token, repo)[:]...)
	}
	list = unique(list)
	if tag != "" {
		listByTag, err := db.Tag(tag)
		log.Check(log.DebugLevel, "Looking for artifacts with tag "+tag, err)
		list = intersect(list, listByTag)
	}
	pstr := strings.Split(page, ",")
	p[0], _ = strconv.Atoi(pstr[0])
	if len(pstr) == 2 {
		p[1], _ = strconv.Atoi(pstr[1])
	}
	log.Debug(fmt.Sprintf("list to be checked: %+v", list))
	for i, k := range list {
		log.Debug(fmt.Sprintf("checking file #%+v: %+v", i, k))
		if (!db.IsPublic(k) && !db.CheckShare(k, db.TokenOwner(token))) ||
			(db.IsPublic(k) && len(owner) > 0 && db.CheckRepo(owner, []string{repo}, k) == 0) ||
			db.CheckRepo("", []string{repo}, k) == 0 {
			log.Debug(fmt.Sprintf("File %+v (name: %+v, owner: %+v, token: %+v) is ignored: %+v || %+v || %+v", k, db.NameByHash(k), owner, db.TokenOwner(token), !db.IsPublic(k) && !db.CheckShare(k, db.TokenOwner(token)), db.IsPublic(k) && len(owner) > 0 && db.CheckRepo(owner, []string{repo}, k) == 0, db.CheckRepo("", []string{repo}, k) == 0))
			continue
		}
		if p[0]--; p[0] > 0 {
			continue
		}
		item := FormatItem(db.Info(k), repo)
		log.Debug(fmt.Sprintf("File #%+v (hash: %+v) in formatted way: %+v", i, k, item))
		if (name == "" || (name != "" && (name == item.Name /*|| (repo == "template" && strings.HasPrefix(item.Name, name) )*/))) &&
			(version == "" || (version != "" && item.Version == version)) {
			items = append(items, item)
		}
		if len(items) >= p[1] {
			break
		}
	}
	log.Debug(fmt.Sprintf("*** Final list of items: %+v", items))
	output, err := json.Marshal(items)
	if err != nil {
		return nil
	}
	if string(output) == "null" {
		output = []byte("[]")
	}
	return output
}

func processVersion(version string) string {
	if version == "latest" {
		return ""
	}
	return version
}

func In(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

func GetVerified(list []string, name, repo, versionTemplate string) ListItem {
	log.Debug(fmt.Sprintf("Getting file \"%+v\" from verified users", name))
	latestVersion, _ := semver.Make("")
	var itemLatestVersion ListItem
	log.Debug(fmt.Sprintf("Iterating through list:\n["))
	for _, k := range list {
		log.Debug(fmt.Sprintf("------------- %+v (name: %+v)", k, db.NameByHash(k)))
	}
	log.Debug(fmt.Sprintf("\n]"))
	for _, k := range list {
		if info := db.Info(k); db.CheckRepo("", []string{repo}, k) > 0 {
			log.Debug(fmt.Sprintf("info[\"name\"] %+v == %+v name (%+v)", info["name"], name, info["name"] == name))
			if info["name"] == name || (strings.HasPrefix(info["name"], name+"-subutai-template") && repo == "template") {
				for _, owner := range db.FileField(info["id"], "owner") {
					itemVersion, _ := semver.Make(info["version"])
					if In(owner, []string{"subutai", "jenkins", "docker", "travis", "appveyor", "devops"}) {
						if itemVersion.GTE(latestVersion) && len(versionTemplate) == 0 {
							log.Debug(fmt.Sprintf("First if %+v", k))
							latestVersion = itemVersion
							itemLatestVersion = FormatItem(db.Info(k), repo)
						} else if versionTemplate == itemVersion.String() {
							log.Debug(fmt.Sprintf("Second if %+v", k))
							itemLatestVersion = FormatItem(db.Info(k), repo)
						}
					}
				}
			}
		}
	}
	return itemLatestVersion
}

func FormatItem(info map[string]string, repo string) ListItem {
	log.Debug(fmt.Sprintf("Repo: %+v, formatting item %+v", repo, info))
	if len(info["prefsize"]) == 0 && repo == "template" {
		info["prefsize"] = "tiny"
	}
	date, _ := time.Parse(time.RFC3339Nano, info["date"])
	timestamp := strconv.FormatInt(date.Unix(), 10)
	item := ListItem{
		ID:            info["id"],
		Date:          date,
		Hash:          hashsums{Md5: info["md5"], Sha256: info["sha256"]},
		Name:          strings.Split(info["name"], "-subutai-template")[0],
		Tags:          db.FileField(info["id"], "tag"),
		Owner:         db.FileField(info["id"], "owner"),
		Version:       info["version"],
		Filename:      info["name"],
		Parent:        info["parent"],
		ParentVersion: info["parent-version"],
		ParentOwner:   info["parent-owner"],
		Prefsize:      info["prefsize"],
		Architecture:  strings.ToUpper(info["arch"]),
		Description:   info["Description"],
		Timestamp:     timestamp,
	}
	item.Size, _ = strconv.Atoi(info["size"])
	if repo == "apt" {
		item.Version = info["Version"]
		item.Architecture = info["Architecture"]
		item.Size, _ = strconv.Atoi(info["Size"])
		item.Hash.Sha256 = info["SHA256"]
	}
	if len(item.Hash.Md5) == 0 {
		item.Hash.Md5 = item.ID
	}
	log.Debug(fmt.Sprintf("Repo: %+v, item after formatting: %+v", repo, item))
	return item
}

func intersect(listA, listB []string) (list []string) {
	mapA := make(map[string]bool)
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

func unique(list []string) []string {
	was := make(map[string]bool)
	result := []string{}
	for _, v := range list {
		if !was[v] {
			was[v] = true
			result = append(result, v)
		}
	}
	return result
}
