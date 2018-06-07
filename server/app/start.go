package app

import (
	"net/http"
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"context"
)

var (
	Server     *http.Server
	Start      chan bool
	Stop       chan bool
	Registered bool
)

func RunServer() {
	Start = make(chan bool)
	Stop = nil
	go ListenAndServe()
	<-Start
	close(Start)
	go WaitShutdown()
}

func ListenAndServe() {
	defer db.Close()
	if !Registered {
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
		Registered = true
	}
	Server = &http.Server{
		Addr:    ":" + config.Network.Port,
		Handler: nil,
	}
	log.Info("Server has started. " + "Listening at " + "http://127.0.0.1:8080")
	Start <- true
	Server.ListenAndServe()
	Server.Close()
}

func WaitShutdown() {
	Stop = make(chan bool)
	log.Info("Waiting for shut down request...")
	loop:
	for {
		select {
		case <-Stop: {
			log.Info("Received shut down request. Stopping server...")
			ctx := context.Background()
			Server.Shutdown(ctx)
			break loop
		}
		}
	}
	log.Info("Server stopped successfully")
	Stop <- true
}