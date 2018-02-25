package apt

import (
	"archive/tar"
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
	"github.com/subutai-io/agent/log"
	"os/exec"
)

func readDeb(hash string) (control bytes.Buffer, err error) {
	file, err := os.Open(config.Storage.Path + hash)
	log.Check(log.WarnLevel, "Opening deb package", err)

	defer file.Close()

	library := ar.NewReader(file)
	for header, err := library.Next(); err != io.EOF; header, err = library.Next() {
		if err != nil {
			return control, err
		}
		if header.Name == "control.tar.gz" {
			ungzip, err := gzip.NewReader(library)
			if err != nil {
				return control, err
			}

			defer ungzip.Close()

			tr := tar.NewReader(ungzip)
			for tarHeader, err := tr.Next(); err != io.EOF; tarHeader, err = tr.Next() {
				if err != nil {
					return control, err
				}
				if tarHeader.Name == "./control" {
					if _, err := io.Copy(&control, tr); err != nil {
						return control, err
					}
					break
				}
			}
		}
	}
	return
}

func getControl(control bytes.Buffer) map[string]string {
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

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		_, header, _ := r.FormFile("file")
		md5, sha256, owner := upload.Handler(w, r)
		if len(md5) == 0 || len(sha256) == 0 {
			return
		}
		control, err := readDeb(md5)
		if err != nil {
			log.Warn(err.Error())
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte(err.Error()))
			if db.Delete(owner, "apt", md5) == 0 {
				os.Remove(config.Storage.Path + md5)
			}
			return
		}
		meta := getControl(control)
		meta["Filename"] = header.Filename
		meta["Size"] = getSize(config.Storage.Path + md5)
		meta["SHA512"] = upload.Hash(config.Storage.Path+md5, "sha512")
		meta["SHA256"] = upload.Hash(config.Storage.Path+md5, "sha256")
		meta["SHA1"] = upload.Hash(config.Storage.Path+md5, "sha1")
		meta["MD5sum"] = md5
		meta["type"] = "apt"
		db.Write(owner, md5, header.Filename, meta)
		w.Write([]byte(md5))
		log.Info(meta["Filename"] + " saved to apt repo by " + owner)
		os.Rename(config.Storage.Path+md5, config.Storage.Path+header.Filename)
		renameOldDebFiles()
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("hash")
	if len(file) == 0 {
		file = strings.TrimPrefix(r.RequestURI, "/kurjun/rest/apt/")
	}
	if f, err := os.Open(config.Storage.Path + file); err == nil && file != "" {
		defer f.Close()
		io.Copy(w, f)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		if hash := upload.Delete(w, r); len(hash) != 0 {
			w.Write([]byte("Removed"))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
	}
}

func Info(w http.ResponseWriter, r *http.Request) {
	if info := download.Info("apt", r); len(info) != 0 {
		w.Write(info)
		return
	}
	w.Write([]byte("Not found"))
}

func renameOldDebFiles()  {
	list := db.Search("")
	for _, k := range list {
		if db.CheckRepo("", "apt", k) == 0 {
			continue
		}
		item := download.FormatItem(db.Info(k), "apt", "")
		os.Rename(config.Storage.Path+item.Hash.Md5, config.Storage.Path+item.Name)
	}
}

func GenerateReleaseFile()  {
	cmd := exec.Command("bash", "-c", "dpkg-scanpackages . /dev/null | tee Packages | gzip > Packages.gz")
	cmd.Dir = config.Storage.Path
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
		log.Info("Can't run dpkg-scanpackages")
	}

	cmd = exec.Command("bash", "-c", "apt-ftparchive release . > Release")
	cmd.Dir = config.Storage.Path
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		log.Info("Can't run apt-ftparchive")
	}
	cmd = exec.Command("bash", "-c", "gpg --batch --yes --armor -u subutai-release@subut.ai -abs -o Release.gpg Release")
	cmd.Dir = config.Storage.Path
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		log.Info("Can't sign Realease file")
	}
}