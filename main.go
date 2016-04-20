package main

import (
	"net/http"

	"github.com/optdyn/gorjun/apt"
	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/raw"
	"github.com/optdyn/gorjun/template"
)

func main() {
	defer db.Close()
	http.HandleFunc("/template", template.Show)
	http.HandleFunc("/template/upload", template.Upload)
	http.HandleFunc("/template/search", template.Search)
	http.HandleFunc("/template/download", template.Download)
	http.HandleFunc("/raw", raw.Show)
	http.HandleFunc("/raw/upload", raw.Upload)
	http.HandleFunc("/raw/download", raw.Download)
	http.HandleFunc("/apt/upload", apt.Upload)
	http.HandleFunc("/apt/", apt.Download)
	http.ListenAndServe(":8080", nil)
}
