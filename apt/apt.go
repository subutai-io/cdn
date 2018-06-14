package apt

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/cdn/download"
	"github.com/subutai-io/cdn/upload"

	"os/exec"

	"github.com/mkrautz/goar"
	"github.com/satori/go.uuid"
	"github.com/subutai-io/agent/log"
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

func getSize(file string) (size int) {
	f, err := os.Open(file)
	if !log.Check(log.WarnLevel, "Opening file "+file, err) {
		stat, _ := f.Stat()
		f.Close()
		size = int(stat.Size())
	}
	return size
}

func Upload(w http.ResponseWriter, r *http.Request) {
	log.Info("Start uploading the file")
	if r.Method == "POST" {
		_, header, _ := r.FormFile("file")
		md5, sha256, owner := upload.Handler(w, r)
		if len(md5) == 0 || len(sha256) == 0 {
			log.Info("Md5 or sha256 is empty. Failed to calculate the hash")
			return
		}
		log.Info(fmt.Sprintf("Starting to read deb package %v", header.Filename))
		control, err := readDeb(header.Filename)
		if err != nil {
			log.Warn("Reading deb package finished with error: ", err.Error())
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte(err.Error()))
			if db.Delete(owner, "apt", md5) == 0 {
				log.Info(fmt.Sprintf("Removed file %v", config.Storage.Path+md5))
				os.Remove(config.Storage.Path + md5)
			}
			return
		}
		log.Info("Starting to read control file of deb package")
		meta := getControl(control)
		meta["Filename"] = header.Filename
		meta["Size"] = strconv.Itoa(getSize(config.Storage.Path + header.Filename))
		meta["SHA512"] = upload.Hash(config.Storage.Path+header.Filename, "sha512")
		meta["SHA256"] = upload.Hash(config.Storage.Path+header.Filename, "sha256")
		meta["SHA1"] = upload.Hash(config.Storage.Path+header.Filename, "sha1")
		meta["md5"] = md5
		meta["type"] = "apt"
		tags := r.FormValue("tag")
		meta["tag"] = tags
		my_uuid, err := uuid.NewV4()
		if err != nil {
			log.Warn("Failed to generate UUID")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		ID := my_uuid.String()
		db.AddTag(strings.Split(tags, ","), ID, "apt")
		log.Info(fmt.Sprintf("Writing deb package %v into database", header.Filename))
		err = db.Write(owner, ID, header.Filename, meta)
		if err != nil {
			log.Warn("Failed writing record into database")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(r.MultipartForm.Value["private"]) > 0 && r.MultipartForm.Value["private"][0] == "true" {
			log.Info("Sharing " + ID + " with " + owner)
			db.MakePrivate(ID, owner)
		} else {
			db.MakePublic(ID, owner)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ID))
		log.Info(meta["Filename"] + " saved to apt repo by " + owner)
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("hash")
	log.Info(fmt.Sprintf("Starting download the deb package %v", file))
	if len(file) == 0 {
		file = strings.TrimPrefix(r.RequestURI, "/kurjun/rest/apt/")
	}
	size := getSize(config.Storage.Path + "Packages")
	if file == "Packages" && size == 0 {
		GenerateReleaseFile()
	}
	log.Info(fmt.Sprintf("Opening file %v", config.Storage.Path+file))
	if f, err := os.Open(config.Storage.Path + file); err == nil && file != "" {
		defer f.Close()
		io.Copy(w, f)
	} else {
		log.Info(fmt.Sprintf("File %v not found", config.Storage.Path+file))
		w.WriteHeader(http.StatusNotFound)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		if hash := upload.Delete(w, r); len(hash) != 0 {
			log.Info(fmt.Sprintf("Removed file with hash %v", hash))
			w.Write([]byte("Removed"))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
	}
}

func Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		return
	}
	if info := download.Info("apt", r); len(info) != 0 {
		w.Write(info)
		return
	}
	w.Write([]byte("Not found"))
}

func List(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		return
	}
	w.Write(download.List("apt", r))
}

func renameOldDebFiles() {
	log.Info("Starting rename old deb files")
	list := db.SearchName("")
	for _, k := range list {
		if db.CheckRepo("", []string{"apt"}, k) == 0 {
			continue
		}
		item := download.FormatItem(db.Info(k), "apt")
		os.Rename(config.Storage.Path+item.Hash.Md5, config.Storage.Path+item.Name)
	}
	log.Info("Finished rename old deb packages")
}

func GenerateReleaseFile() {
	log.Warn("Starting GenerateReleaseFiles")
	cmd := exec.Command("bash", "-c", "dpkg-scanpackages . /dev/null | tee Packages | gzip > Packages.gz")
	cmd.Dir = config.Storage.Path
	log.Warn("config.Storage.Path: ", config.Storage.Path)
	log.Warn("Cmd.Dir: ", cmd.Dir)
	log.Warn("Running command to run dpkg-scanpackages and waiting for it to finish")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
		log.Info("Can't run dpkg-scanpackages")
	}
	log.Warn(fmt.Sprintf("Command to run dpkg-scanpackages finished with error %v", err))

	cmd = exec.Command("bash", "-c", "apt-ftparchive release . > Release")
	cmd.Dir = config.Storage.Path
	log.Warn("config.Storage.Path: ", config.Storage.Path)
	log.Warn("Cmd.Dir: ", cmd.Dir)
	log.Warn("Running command to run apt-ftparchive and waiting for it to finish")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		log.Info("Can't run apt-ftparchive")
	}
	log.Warn(fmt.Sprintf("Command to run apt-ftparchive finished with error %v", err))

	cmd = exec.Command("bash", "-c", "gpg --batch --yes --armor -u subutai-release@subutai.io -abs -o Release.gpg Release")
	cmd.Dir = config.Storage.Path
	log.Warn("config.Storage.Path: ", config.Storage.Path)
	log.Warn("Cmd.Dir: ", cmd.Dir)
	log.Warn("Running command to sign Realease file and waiting for it to finish")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		log.Info("Can't sign Realease file")
	}
	log.Warn(fmt.Sprintf("Command to sign Realease file finished with error %v", err))
	log.Warn("Finished GenerateReleaseFiles")
}

func Generate(w http.ResponseWriter, r *http.Request) {
	log.Info("Starting Generate")
	token := strings.ToLower(r.Header.Get("token"))
	owner := strings.ToLower(db.TokenOwner(token))
	if len(token) == 0 || len(owner) == 0 {
		log.Warn("Not authorized user wanted to generate release file")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not authorized"))
		log.Warn(r.RemoteAddr + " - rejecting generate request")
		return
	}
	if owner != "subutai" && owner != "jenkins" {
		log.Warn("Not allowed user wanted to generate release file")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only allowed users can generate release file"))
		log.Warn(r.RemoteAddr + " - rejecting generate request")
		return
	}
	log.Info("Generating release file")
	GenerateReleaseFile()
	w.Write([]byte("New Packages file generated and Release file signed"))
	log.Info("New Packages file generated and Release file signed")
}
