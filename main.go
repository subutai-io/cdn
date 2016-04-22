package main

import (
	"net/http"

	"github.com/optdyn/gorjun/apt"
	"github.com/optdyn/gorjun/auth"
	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/raw"
	"github.com/optdyn/gorjun/template"
)

func main() {
	defer db.Close()
	http.HandleFunc("/kurjun/rest/template", template.Show)
	http.HandleFunc("/kurjun/rest/template/info", template.Info)
	http.HandleFunc("/kurjun/rest/template/upload", template.Upload)
	http.HandleFunc("/kurjun/rest/template/search", template.Search)
	http.HandleFunc("/kurjun/rest/template/download", template.Download)
	http.HandleFunc("/kurjun/rest/template/delete", template.Delete)
	http.HandleFunc("/kurjun/rest/template/get", template.Download)
	http.HandleFunc("/kurjun/rest/template/list", template.List)
	http.HandleFunc("/kurjun/rest/deb/list", template.List)
	http.HandleFunc("/kurjun/rest/file/list", template.List)
	http.HandleFunc("/kurjun/rest/template/md5", template.Md5)
	http.HandleFunc("/kurjun/rest/deb/md5", template.Md5)
	http.HandleFunc("/kurjun/rest/file/md5", template.Md5)
	http.HandleFunc("/kurjun/rest/raw", raw.Show)
	http.HandleFunc("/kurjun/rest/raw/upload", raw.Upload)
	http.HandleFunc("/kurjun/rest/raw/download", raw.Download)
	http.HandleFunc("/kurjun/rest/apt/delete", apt.Delete)
	http.HandleFunc("/kurjun/rest/apt/upload", apt.Upload)
	http.HandleFunc("/kurjun/rest/apt/", apt.Download)
	http.HandleFunc("/kurjun/rest/auth/register", auth.Register)
	http.HandleFunc("/kurjun/rest/auth/token", auth.Token)
	http.ListenAndServe(":8080", nil)
}
