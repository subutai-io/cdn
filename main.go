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
	http.HandleFunc("/kurjun/rest/template/get", template.Download)
	http.HandleFunc("/raw", raw.Show)
	http.HandleFunc("/raw/upload", raw.Upload)
	http.HandleFunc("/raw/download", raw.Download)
	http.HandleFunc("/apt/upload", apt.Upload)
	http.HandleFunc("/apt/", apt.Download)
	http.HandleFunc("/auth/register", auth.Register)
	http.HandleFunc("/auth/token", auth.Token)
	// http.HandleFunc("/pgp/verify", pgp.Verify)
	http.ListenAndServe(":8080", nil)
}
