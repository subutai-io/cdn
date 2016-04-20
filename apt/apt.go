package apt

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strconv"
	// "path/filepath"
	"strings"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/upload"

	"github.com/mkrautz/goar"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/deb/"
)

func readDeb(hash string) (string, bytes.Buffer) {
	var control bytes.Buffer
	file, err := os.Open(path + hash)
	log.Check(log.FatalLevel, "Opening deb package", err)

	defer file.Close()

	library := ar.NewReader(file)
	for {
		header, err := library.Next()
		if err == io.EOF {
			break
		}
		log.Check(log.FatalLevel, "Reading deb content", err)

		if header.Name == "control.tar.gz" {
			ungzip, err := gzip.NewReader(library)
			log.Check(log.FatalLevel, "Ungziping control file", err)

			defer ungzip.Close()

			tr := tar.NewReader(ungzip)
			for {
				tarHeader, err := tr.Next()
				if err == io.EOF {
					break
				}
				log.Check(log.FatalLevel, "Reading control tar", err)

				if tarHeader.Name == "./control" {
					if _, err := io.Copy(&control, tr); err != nil {
						log.Fatal(err.Error())
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
	log.Check(log.FatalLevel, "Opening file "+file, err)
	defer f.Close()

	stat, err := f.Stat()
	log.Check(log.FatalLevel, "Getting file stat", err)
	return strconv.Itoa(int(stat.Size()))
}

func writePackage(meta map[string]string) {
	var f *os.File
	if _, err := os.Stat(path + "Packages"); os.IsNotExist(err) {
		f, err = os.Create(path + "Packages")
		log.Check(log.FatalLevel, "Creating packages file", err)
		defer f.Close()
	} else if err == nil {
		f, err = os.OpenFile(path+"Packages", os.O_APPEND|os.O_WRONLY, 0600)
		log.Check(log.FatalLevel, "Opening packages file", err)
		defer f.Close()
	} else {
		log.Fatal(err.Error())
	}

	for k, v := range meta {
		_, err := f.WriteString(string(k) + ": " + string(v) + "\n")
		log.Check(log.FatalLevel, "Appending package data", err)
	}
	_, err := f.Write([]byte("\n"))
	log.Check(log.FatalLevel, "Appending endline", err)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("apt")))
	} else if r.Method == "POST" {
		_, header, _ := r.FormFile("file")
		hash, meta := getControl(readDeb(upload.Handler(w, r)))
		meta["Filename"] = header.Filename
		meta["Size"] = getSize(path + hash)
		writePackage(meta)
		w.Write([]byte("Name: " + header.Filename + "\n"))
		db.Write(hash, header.Filename, meta)
		w.Write([]byte("Added to db: " + db.Read(hash)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimLeft(r.RequestURI, "/apt/")
	log.Info("Request: " + r.RequestURI)
	if file != "Packages.gz" && file != "InRelease" && file != "Release" {
		for k, _ := range db.Search(file) {
			file = k
		}
	}
	f, err := os.Open(path + file)
	if log.Check(log.WarnLevel, "Opening file "+path+file, err) {
		w.WriteHeader(http.StatusNotFound)
	}
	defer f.Close()
	io.Copy(w, f)
}
