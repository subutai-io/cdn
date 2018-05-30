package main

import (
	"net/http"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/server/app"
)

var (
	srv *http.Server
)

func init() {
	app.InitFilters()
}

// main starts CDN server
func main() {
	defer db.Close()

	log.Info("Server has started. " + "Listening at " + "127.0.0.1:8080")

	http.HandleFunc("/kurjun/rest/auth/key", app.Key)
	http.HandleFunc("/kurjun/rest/auth/keys", app.Keys)
	http.HandleFunc("/kurjun/rest/auth/sign", app.Sign)
	http.HandleFunc("/kurjun/rest/auth/owner", app.Owner)
	http.HandleFunc("/kurjun/rest/auth/token", app.Token)
	http.HandleFunc("/kurjun/rest/auth/register", app.Register)
	http.HandleFunc("/kurjun/rest/auth/validate", app.Validate)

	http.HandleFunc("/kurjun/rest/apt/info", app.FileSearch)
	http.HandleFunc("/kurjun/rest/apt/list", app.FileSearch)
	http.HandleFunc("/kurjun/rest/apt/upload", app.FileUpload)

	http.HandleFunc("/kurjun/rest/raw/info", app.FileSearch)
	http.HandleFunc("/kurjun/rest/raw/list", app.FileSearch)
	http.HandleFunc("/kurjun/rest/raw/upload", app.FileUpload)

	http.HandleFunc("/kurjun/rest/template/info", app.FileSearch)
	http.HandleFunc("/kurjun/rest/template/list", app.FileSearch)
	http.HandleFunc("/kurjun/rest/template/upload", app.FileUpload)

	srv = &http.Server{
		Addr:    ":" + config.Network.Port,
		Handler: nil,
	}

	srv.ListenAndServe()
}
