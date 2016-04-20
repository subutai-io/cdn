package apt

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/upload"

	"github.com/mkrautz/goar"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

func compress(source string) {
	reader, err := os.Open(source)
	log.Check(log.FatalLevel, "Opening file "+source+" to compress", err)

	writer, err := os.Create(source + ".gz")
	log.Check(log.FatalLevel, "Creating gz target", err)
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filepath.Base(source)
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	log.Check(log.FatalLevel, "Writing data to archive", err)
}

func decompress(source string) {
	reader, err := os.Open(source)
	log.Check(log.FatalLevel, "Opening file "+source+" to decompress", err)
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	log.Check(log.FatalLevel, "Creating gz reader", err)
	defer archive.Close()

	writer, err := os.Create(strings.TrimRight(source, ".gz"))
	log.Check(log.FatalLevel, "Creating write target", err)
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	log.Check(log.FatalLevel, "Writing archive data to file", err)
}

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

func writePackage(meta map[string]string) {
	var f *os.File
	if _, err := os.Stat(path + "Packages.gz"); os.IsNotExist(err) {
		f, err = os.Create(path + "Packages")
		log.Check(log.FatalLevel, "Creating packages file", err)
		defer f.Close()
	} else if err == nil {
		decompress(path + "Packages.gz")
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
	os.Remove(path + "Packages.gz")
	compress(path + "Packages")
	os.Remove(path + "Packages")
	return
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(upload.Page("apt")))
	} else if r.Method == "POST" {
		hash, meta := getControl(readDeb(upload.Handler(w, r)))
		writePackage(meta)
		_, header, _ := r.FormFile("file")
		w.Write([]byte("Name: " + header.Filename + "\n"))
		db.Write(header.Filename, hash, meta)
		w.Write([]byte("Added to db: " + db.Read(header.Filename)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	file := strings.TrimLeft(r.RequestURI, "/apt/")
	if file != "Packages.gz" {
		file = db.Read(file)
		w.Write([]byte(file))
	}
	f, err := os.Open(path + file)
	log.Check(log.FatalLevel, "Opening file "+path+file, err)
	defer f.Close()
	io.Copy(w, f)
}
