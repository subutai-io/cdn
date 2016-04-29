package main

import (
	"net/http"

	"github.com/subutai-io/gorjun/apt"
	"github.com/subutai-io/gorjun/auth"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/raw"
	"github.com/subutai-io/gorjun/template"
)

func main() {
	defer db.Close()
	http.HandleFunc("/kurjun/rest/template/info", template.Info)
	http.HandleFunc("/kurjun/rest/template/upload", template.Upload)
	http.HandleFunc("/kurjun/rest/template/download", template.Download)
	http.HandleFunc("/kurjun/rest/template/delete", template.Delete)
	http.HandleFunc("/kurjun/rest/template/get", template.Download)
	http.HandleFunc("/kurjun/rest/template/list", template.List)
	http.HandleFunc("/kurjun/rest/template/md5", template.Md5)

	http.HandleFunc("/kurjun/rest/raw/upload", raw.Upload)
	http.HandleFunc("/kurjun/rest/raw/download", raw.Download)
	http.HandleFunc("/kurjun/rest/raw/delete", raw.Delete)
	http.HandleFunc("/kurjun/rest/raw/info", raw.Info)
	http.HandleFunc("/kurjun/rest/raw/list", raw.List)
	http.HandleFunc("/kurjun/rest/file/info", raw.Info)
	http.HandleFunc("/kurjun/rest/file/list", raw.List)
	http.HandleFunc("/kurjun/rest/file/md5", template.Md5)
	http.HandleFunc("/kurjun/rest/file/get", raw.Download)
	http.HandleFunc("/kurjun/rest/file/delete", raw.Delete)

	http.HandleFunc("/kurjun/rest/apt/", apt.Download)
	http.HandleFunc("/kurjun/rest/apt/info", apt.Info)
	http.HandleFunc("/kurjun/rest/apt/list", apt.Info)
	http.HandleFunc("/kurjun/rest/apt/delete", apt.Delete)
	http.HandleFunc("/kurjun/rest/apt/upload", apt.Upload)
	http.HandleFunc("/kurjun/rest/apt/download", apt.Download)
	http.HandleFunc("/kurjun/rest/deb/md5", template.Md5)
	http.HandleFunc("/kurjun/rest/deb/list", apt.Info)

	http.HandleFunc("/kurjun/rest/auth/register", auth.Register)
	http.HandleFunc("/kurjun/rest/auth/validate", auth.Validate)
	http.HandleFunc("/kurjun/rest/auth/token", auth.Token)
	http.ListenAndServe(":8080", nil)
}
