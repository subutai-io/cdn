package apt

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"

	"github.com/mkrautz/goar"
	"github.com/subutai-io/base/agent/log"
)

func readDeb(hash string) (string, bytes.Buffer) {
	var control bytes.Buffer
	file, err := os.Open(config.Storage.Path + hash)
	log.Check(log.WarnLevel, "Opening deb package", err)

	defer file.Close()

	library := ar.NewReader(file)
	for header, err := library.Next(); err != io.EOF; header, err = library.Next() {
		if header.Name == "control.tar.gz" {
			ungzip, err := gzip.NewReader(library)
			log.Check(log.WarnLevel, "Ungziping control file", err)

			defer ungzip.Close()

			tr := tar.NewReader(ungzip)
			for tarHeader, err := tr.Next(); err != io.EOF; tarHeader, err = tr.Next() {
				if tarHeader.Name == "./control" {
					if _, err := io.Copy(&control, tr); err != nil {
						log.Warn(err.Error())
					}
					break
				}
			}
		}
	}
	return hash, control
}

func getControl(hash string, control bytes.Buffer) map[string]string {
	d := make(map[string]string)
	for _, v := range strings.Split(control.String(), "\n") {
		line := strings.Split(v, ":")
		if len(line) > 1 {
			d[line[0]] = strings.TrimPrefix(line[1], " ")
		}
	}
	return d
}

func getSize(file string) (size string) {
	f, err := os.Open(file)
	if !log.Check(log.WarnLevel, "Opening file "+file, err) {
		stat, _ := f.Stat()
		f.Close()
		size = strconv.Itoa(int(stat.Size()))
	}
	return size
}

func writePackage(meta map[string]string) {
	var f *os.File
	if _, err := os.Stat(config.Storage.Path + "Packages"); os.IsNotExist(err) {
		f, err = os.Create(config.Storage.Path + "Packages")
		log.Check(log.WarnLevel, "Creating packages file", err)
		defer f.Close()
	} else if err == nil {
		f, err = os.OpenFile(config.Storage.Path+"Packages", os.O_APPEND|os.O_WRONLY, 0600)
		log.Check(log.WarnLevel, "Opening packages file", err)
		defer f.Close()
	} else {
		log.Warn(err.Error())
	}

	for k, v := range meta {
		_, err := f.WriteString(string(k) + ": " + string(v) + "\n")
		log.Check(log.WarnLevel, "Appending package data", err)
	}
	_, err := f.Write([]byte("\n"))
	log.Check(log.WarnLevel, "Appending endline", err)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		_, header, _ := r.FormFile("file")
		hash, owner := upload.Handler(w, r)
		meta := getControl(readDeb(hash))
		meta["Filename"] = header.Filename
		meta["Size"] = getSize(config.Storage.Path + hash)
		meta["MD5sum"] = hash
		meta["type"] = "apt"
		writePackage(meta)
		db.Write(owner, hash, header.Filename, meta)
		w.Write([]byte(hash))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("hash")
	if len(file) == 0 {
		file = strings.TrimPrefix(r.RequestURI, "/kurjun/rest/apt/")
	}
	if file != "Packages" && file != "InRelease" && file != "Release" {
		file = db.LastHash(file, "apt")
	}
	f, err := os.Open(config.Storage.Path + file)
	if log.Check(log.WarnLevel, "Opening file "+config.Storage.Path+file, err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func readPackages() []string {
	file, err := os.Open(config.Storage.Path + "Packages")
	log.Check(log.WarnLevel, "Opening packages file", err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	log.Check(log.WarnLevel, "Scanning packages list", scanner.Err())
	return lines
}

func deleteInfo(hash string) {
	list := readPackages()
	if len(list) == 0 {
		log.Warn("Empty packages list")
		return
	}

	var newlist, block string
	changed, skip := false, false
	for _, line := range list {
		if len(line) != 0 && skip {
			continue
		} else if len(line) == 0 {
			skip = false
			if len(block) != 0 {
				newlist = newlist + block + "\n"
				block = ""
			}
		} else if len(line) != 0 && !skip {
			if strings.HasSuffix(line, hash) {
				block = ""
				skip = true
				changed = true
			} else {
				block = block + line + "\n"
			}
		}
	}
	if changed {
		log.Info("Updating packages list")
		file, err := os.Create(config.Storage.Path + "Packages.new")
		log.Check(log.WarnLevel, "Opening packages file", err)
		defer file.Close()

		_, err = file.WriteString(newlist)
		log.Check(log.WarnLevel, "Writing new list", err)
		log.Check(log.WarnLevel, "Replacing old list",
			os.Rename(config.Storage.Path+"Packages.new", config.Storage.Path+"Packages"))
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if hash := upload.Delete(w, r); len(hash) != 0 && r.Method == "DELETE" {
		deleteInfo(hash)
		w.Write([]byte("Removed"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Incorrect method"))
}

func Info(w http.ResponseWriter, r *http.Request) {
	if info := download.Info("apt", r); len(info) != 0 {
		w.Write(info)
		return
	}
	if js := download.ProxyList("apt"); js != nil {
		w.Write(js)
		return
	}
	w.Write([]byte("Not found"))
}
