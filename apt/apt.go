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

	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/download"
	"github.com/subutai-io/gorjun/upload"

	"github.com/mkrautz/goar"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

func readDeb(hash string) (string, bytes.Buffer) {
	var control bytes.Buffer
	file, err := os.Open(path + hash)
	log.Check(log.WarnLevel, "Opening deb package", err)

	defer file.Close()

	library := ar.NewReader(file)
	for {
		header, err := library.Next()
		if err == io.EOF {
			break
		}
		log.Check(log.WarnLevel, "Reading deb content", err)

		if header.Name == "control.tar.gz" {
			ungzip, err := gzip.NewReader(library)
			log.Check(log.WarnLevel, "Ungziping control file", err)

			defer ungzip.Close()

			tr := tar.NewReader(ungzip)
			for {
				tarHeader, err := tr.Next()
				if err == io.EOF {
					break
				}
				log.Check(log.WarnLevel, "Reading control tar", err)

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

func getControl(hash string, control bytes.Buffer) (string, map[string]string) {
	d := make(map[string]string)
	for _, v := range strings.Split(control.String(), "\n") {
		line := strings.Split(v, ":")
		if len(line) == 2 {
			d[line[0]] = strings.TrimLeft(line[1], " ")
		}
	}
	return hash, d
}

func getSize(file string) string {
	f, err := os.Open(file)
	log.Check(log.WarnLevel, "Opening file "+file, err)
	defer f.Close()

	stat, err := f.Stat()
	log.Check(log.WarnLevel, "Getting file stat", err)
	return strconv.Itoa(int(stat.Size()))
}

func writePackage(meta map[string]string) {
	var f *os.File
	if _, err := os.Stat(path + "Packages"); os.IsNotExist(err) {
		f, err = os.Create(path + "Packages")
		log.Check(log.WarnLevel, "Creating packages file", err)
		defer f.Close()
	} else if err == nil {
		f, err = os.OpenFile(path+"Packages", os.O_APPEND|os.O_WRONLY, 0600)
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
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("apt")))
	} else if r.Method == "POST" {
		_, header, _ := r.FormFile("file")
		hash, owner := upload.Handler(w, r)
		hash, meta := getControl(readDeb(hash))
		meta["Filename"] = header.Filename
		meta["Size"] = getSize(path + hash)
		meta["MD5sum"] = hash
		meta["type"] = "apt"
		writePackage(meta)
		w.Write([]byte("Name: " + header.Filename + "\n"))
		db.Write(owner, hash, header.Filename, meta)
		w.Write([]byte("Added to db: " + db.Read(hash)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("hash")
	if len(file) == 0 {
		file = strings.TrimPrefix(r.RequestURI, "/kurjun/rest/apt/")
		if file != "Packages.gz" && file != "InRelease" && file != "Release" {
			file = db.LastHash(file)
		}
	}
	if len(file) == 0 {
		download.List("apt", w, r)
	} else {
		log.Info("Request: " + file)
		f, err := os.Open(path + file)
		if log.Check(log.WarnLevel, "Opening file "+path+file, err) {
			w.WriteHeader(http.StatusNotFound)
		}
		defer f.Close()
		io.Copy(w, f)
	}
}

func readPackages() []string {
	file, err := os.Open(path + "Packages")
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
		file, err := os.Create(path + "Packages.new")
		log.Check(log.WarnLevel, "Opening packages file", err)
		defer file.Close()

		_, err = file.WriteString(newlist)
		log.Check(log.WarnLevel, "Writing new list", err)
		log.Check(log.WarnLevel, "Replacing old list",
			os.Rename(path+"Packages.new", path+"Packages"))
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		log.Warn("Incorrect method")
		return
	}
	if hash := upload.Delete(w, r); len(hash) != 0 {
		deleteInfo(hash)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Removed"))
	}
}

func Info(w http.ResponseWriter, r *http.Request) {
	info := download.Info("apt", r)
	if len(info) != 0 {
		w.Write(info)
	} else {
		w.Write([]byte("Not found"))
	}
}
