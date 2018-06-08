package app

import (
	"net/http"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"context"
	"github.com/subutai-io/cdn/config"
	"fmt"
	"github.com/subutai-io/cdn/auth"
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
		http.HandleFunc("/kurjun/rest/auth/key", auth.Key)
		http.HandleFunc("/kurjun/rest/auth/keys", auth.Keys)
		http.HandleFunc("/kurjun/rest/auth/sign", auth.Sign)
		http.HandleFunc("/kurjun/rest/auth/owner", auth.Owner)
		http.HandleFunc("/kurjun/rest/auth/token", auth.Token)
		http.HandleFunc("/kurjun/rest/auth/register", auth.Register)
		http.HandleFunc("/kurjun/rest/auth/validate", auth.Validate)
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
	log.Info(fmt.Sprintf("Configuration before starting: %+v %+v %+v %+v", config.ConfigurationCDN, config.ConfigurationDB, config.ConfigurationNetwork, config.ConfigurationStorage))
	Server = &http.Server{
		Addr:    ":" + config.ConfigurationNetwork.Port,
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
			Server.Shutdown(context.Background())
			break loop
		}
		}
	}
	log.Info("Server stopped successfully")
	Stop <- true
}
