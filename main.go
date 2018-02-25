package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/subutai-io/agent/log"

	"github.com/subutai-io/gorjun/apt"
	"github.com/subutai-io/gorjun/auth"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
	"github.com/subutai-io/gorjun/raw"
	"github.com/subutai-io/gorjun/template"
	"github.com/subutai-io/gorjun/upload"
	"github.com/jasonlvhit/gocron"
)

var version = "6.3.0"

var (
	srv      *http.Server
	testMode bool = false
	stop     chan bool
)
func main() {
	defer db.Close()
	// defer torrent.Close()
	// go torrent.SeedLocal()
	gocron.Every(6).Hours().Do(apt.GenerateReleaseFile)
	<- gocron.Start()

	if len(config.CDN.Node) > 0 {
		target := url.URL{Scheme: "https", Host: config.CDN.Node}
		proxy := httputil.NewSingleHostReverseProxy(&target)
		targetQuery := target.RawQuery
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = config.CDN.Node
			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		}
		log.Check(log.ErrorLevel, "Starting to listen :"+config.Network.Port, http.ListenAndServe(":"+config.Network.Port, proxy))
		return
	}

	log.Info("Server has started. " + "Listening at " + "127.0.0.1:8080")

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
	http.HandleFunc("/kurjun/rest/template/tag", template.Tag)
	http.HandleFunc("/kurjun/rest/template/info", template.Info)
	http.HandleFunc("/kurjun/rest/template/list", template.Info)
	http.HandleFunc("/kurjun/rest/template/delete", template.Delete)
	http.HandleFunc("/kurjun/rest/template/upload", template.Upload)
	http.HandleFunc("/kurjun/rest/template/download", template.Download)
	// http.HandleFunc("/kurjun/rest/template/torrent", template.Torrent)

	http.HandleFunc("/kurjun/rest/auth/key", auth.Key)
	http.HandleFunc("/kurjun/rest/auth/keys", auth.Keys)
	http.HandleFunc("/kurjun/rest/auth/sign", auth.Sign)
	http.HandleFunc("/kurjun/rest/auth/token", auth.Token)
	http.HandleFunc("/kurjun/rest/auth/register", auth.Register)
	http.HandleFunc("/kurjun/rest/auth/validate", auth.Validate)

	http.HandleFunc("/kurjun/rest/share", upload.Share)
	http.HandleFunc("/kurjun/rest/quota", upload.Quota)
	http.HandleFunc("/kurjun/rest/about", about)

	if testMode {
		http.HandleFunc("/kurjun/rest/shutdown", shutdown)
	}

	srv = &http.Server{
		Addr:    ":" + config.Network.Port,
		Handler: nil,
	}
	srv.ListenAndServe()
}

func shutdown(w http.ResponseWriter, r *http.Request) {
	log.Info("Shutting down the server")
	stop <- true
}

func about(w http.ResponseWriter, r *http.Request) {
	if strings.Split(r.RemoteAddr, ":")[0] == "127.0.0.1" {
		_, err := w.Write([]byte(version))
		log.Check(log.DebugLevel, "Writing Kurjun version", err)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func runMain() {
	// start the stop channel
	stop = make(chan bool)
	// put the service in "testMode"
	testMode = true
	// run the main entry point
	go main()
	// watch for the stop channel
	<-stop
	// stop the graceful server
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
}
