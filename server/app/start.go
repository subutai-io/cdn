package app

import (
	"net/http"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
)

var (
	Server *http.Server
)

func ListenAndServe() {
	defer db.Close()
	http.HandleFunc("/kurjun/rest/auth/key", Key)
	http.HandleFunc("/kurjun/rest/auth/keys", Keys)
	http.HandleFunc("/kurjun/rest/auth/sign", Sign)
	http.HandleFunc("/kurjun/rest/auth/owner", Owner)
	http.HandleFunc("/kurjun/rest/auth/token", Token)
	http.HandleFunc("/kurjun/rest/auth/register", Register)
	http.HandleFunc("/kurjun/rest/auth/validate", Validate)
	http.HandleFunc("/kurjun/rest/apt/info", FileSearch)
	http.HandleFunc("/kurjun/rest/apt/list", FileSearch)
	http.HandleFunc("/kurjun/rest/apt/upload", FileUpload)
	http.HandleFunc("/kurjun/rest/raw/info", FileSearch)
	http.HandleFunc("/kurjun/rest/raw/list", FileSearch)
	http.HandleFunc("/kurjun/rest/raw/upload", FileUpload)
	http.HandleFunc("/kurjun/rest/template/info", FileSearch)
	http.HandleFunc("/kurjun/rest/template/list", FileSearch)
	http.HandleFunc("/kurjun/rest/template/upload", FileUpload)
	Server = &http.Server{
		Addr:    ":" + config.Network.Port,
		Handler: nil,
	}
	log.Info("Starting server...")
	go Server.ListenAndServe()
	log.Info("Server has started. " + "Listening at " + "http://127.0.0.1:8080")
}
