package main

import (
	"net/http"

	"github.com/optdyn/gorjun/db"
	"github.com/optdyn/gorjun/template"
)

func main() {
	defer db.Close()
	http.HandleFunc("/template", template.List)
	http.HandleFunc("/template/upload", template.Upload)
	http.HandleFunc("/template/download", template.Download)
	http.ListenAndServe(":8080", nil)
}
