package main

import (
	"net/http"

	"github.com/subutai-io/agent/log"

	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/cdn/server/app"
)

var (
	srv *http.Server
)

func main() {
	defer db.Close()

	log.Info("Server has started. " + "Listening at " + "127.0.0.1:8080")

	http.HandleFunc("/kurjun/rest/file/info", app.Info)

	http.HandleFunc("/kurjun/rest/apt/info", app.Info)
	http.HandleFunc("/kurjun/rest/apt/list", app.List)

	http.HandleFunc("/kurjun/rest/raw/info", app.Info)
	http.HandleFunc("/kurjun/rest/raw/list", app.List)

	http.HandleFunc("/kurjun/rest/template/info", app.Info)
	http.HandleFunc("/kurjun/rest/template/list", app.List)

	srv = &http.Server{
		Addr:    ":" + config.Network.Port,
		Handler: nil,
	}

	srv.ListenAndServe()

}
