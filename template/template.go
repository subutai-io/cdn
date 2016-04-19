package template

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/upload"
	"github.com/subutai-io/base/agent/log"
)

var (
	path = "/tmp/"
)

type template struct {
	name, parent, version, arch, hash string
}

var uploadPage string = `
  <html>
  <title>Go upload</title>
  <body>

  <form action="http://localhost:8080/template/upload" method="post" enctype="multipart/form-data">
  <label for="file">Filename:</label>
  <input type="file" name="file" id="file">
  <input type="submit" name="submit" value="Submit">
  </form>

  </body>
  </html>
`

func readTempl(hash string) (string, bytes.Buffer) {
	var config bytes.Buffer
	f, err := os.Open(path + hash)
	log.Check(log.FatalLevel, "Opening file "+path+hash, err)
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	log.Check(log.FatalLevel, "Creating gzip reader", err)

	tr := tar.NewReader(gzf)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		log.Check(log.FatalLevel, "Reading tar content", err)

		if hdr.Name == "config" {
			if _, err := io.Copy(&config, tr); err != nil {
				log.Fatal(err.Error())
			}
			break
		}
	}
	return hash, config
}

func getConf(hash string, config bytes.Buffer) (t *template) {
	t = &template{
		arch:    "lxc.arch",
		name:    "lxc.utsname",
		hash:    hash,
		parent:  "subutai.parent",
		version: "subutai.template.version",
	}

	for _, v := range strings.Split(config.String(), "\n") {
		line := strings.Split(v, "=")
		switch strings.Trim(line[0], " ") {
		case t.arch:
			t.arch = strings.Trim(line[1], " ")
		case t.name:
			t.name = strings.Trim(line[1], " ")
		case t.parent:
			t.parent = strings.Trim(line[1], " ")
		case t.version:
			t.version = strings.Trim(line[1], " ")
		}
	}
	return
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(uploadPage))
	} else if r.Method == "POST" {
		t := getConf(readTempl(upload.Handler(w, r)))
		w.Write([]byte("Name: " + t.name + ", version: " + t.version + ", hash: " + t.hash + "\n"))
		db.Write(t.hash, t.name+"-subutai-template_"+t.version+"_"+t.arch+".tar.gz")
		w.Write([]byte("Added to db: " + db.Read(t.hash)))
	}
}

func Download(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if len(hash) != 0 {
		w.Header().Set("Content-Disposition", "attachment; filename="+db.Read(hash))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		f, err := os.Open(path + hash)
		log.Check(log.FatalLevel, "Opening file "+path+hash, err)
		defer f.Close()
		io.Copy(w, f)
	} else {
		w.Write([]byte("Please specify hash"))
	}
}

func List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html><body>"))
	for k, v := range db.List() {
		// <a href="url">link text</a>
		w.Write([]byte("<p><a href=\"http://localhost:8080/template/download?hash=" + k + "\">" + v + "</a></p>"))
	}
	w.Write([]byte("</body></html>"))
}
