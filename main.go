package main

import (
	"net/http"

	"github.com/subutai-io/gorjun/apt"
	"github.com/subutai-io/gorjun/auth"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/raw"
	"github.com/subutai-io/gorjun/template"
	"github.com/subutai-io/gorjun/upload"
)

func main() {
	defer db.Close()
	// defer torrent.Close()
	// go torrent.SeedLocal()
	http.HandleFunc("/kurjun/rest/file/get", raw.Download)
	http.HandleFunc("/kurjun/rest/file/info", raw.Info)
	http.HandleFunc("/kurjun/rest/raw/get", raw.Download)
	http.HandleFunc("/kurjun/rest/template/get", template.Download)

	http.HandleFunc("/kurjun/rest/apt/", apt.Download)
	http.HandleFunc("/kurjun/rest/apt/info", apt.Info)
	http.HandleFunc("/kurjun/rest/apt/list", apt.Info)
	http.HandleFunc("/kurjun/rest/apt/delete", apt.Delete)
	http.HandleFunc("/kurjun/rest/apt/upload", apt.Upload)
	http.HandleFunc("/kurjun/rest/apt/download", apt.Download)

	http.HandleFunc("/kurjun/rest/raw/", raw.Download)
	http.HandleFunc("/kurjun/rest/raw/info", raw.Info)
	http.HandleFunc("/kurjun/rest/raw/list", raw.Info)
	http.HandleFunc("/kurjun/rest/raw/delete", raw.Delete)
	http.HandleFunc("/kurjun/rest/raw/upload", raw.Upload)
	http.HandleFunc("/kurjun/rest/raw/download", raw.Download)

	http.HandleFunc("/kurjun/rest/template/", template.Download)
	http.HandleFunc("/kurjun/rest/template/info", template.Info)
	http.HandleFunc("/kurjun/rest/template/list", template.Info)
	http.HandleFunc("/kurjun/rest/template/delete", template.Delete)
	http.HandleFunc("/kurjun/rest/template/upload", template.Upload)
	// http.HandleFunc("/kurjun/rest/template/torrent", template.Torrent)
	http.HandleFunc("/kurjun/rest/template/download", template.Download)

	http.HandleFunc("/kurjun/rest/auth/key", auth.Key)
	http.HandleFunc("/kurjun/rest/auth/sign", auth.Sign)
	http.HandleFunc("/kurjun/rest/auth/token", auth.Token)
	http.HandleFunc("/kurjun/rest/auth/register", auth.Register)
	http.HandleFunc("/kurjun/rest/auth/validate", auth.Validate)

	http.HandleFunc("/kurjun/rest/share", upload.Share)
	http.HandleFunc("/kurjun/rest/quota", upload.Quota)

	http.ListenAndServe(":"+config.Network.Port, nil)
}
