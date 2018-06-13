package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/subutai-io/agent/log"
	"github.com/asdine/storm"
)

var (
	DB         *storm.DB
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
	defer DB.Close()
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
		http.HandleFunc("/kurjun/rest/apt/delete", FileDelete)
		http.HandleFunc("/kurjun/rest/apt/", FileDownload)
		http.HandleFunc("/kurjun/rest/apt/download", FileDownload)
		http.HandleFunc("/kurjun/rest/raw/info", FileSearch)
		http.HandleFunc("/kurjun/rest/raw/list", FileSearch)
		http.HandleFunc("/kurjun/rest/raw/upload", FileUpload)
		http.HandleFunc("/kurjun/rest/raw/delete", FileDelete)
		http.HandleFunc("/kurjun/rest/raw/download", FileDownload)
		http.HandleFunc("/kurjun/rest/template/info", FileSearch)
		http.HandleFunc("/kurjun/rest/template/list", FileSearch)
		http.HandleFunc("/kurjun/rest/template/upload", FileUpload)
		http.HandleFunc("/kurjun/rest/template/delete", FileDelete)
		http.HandleFunc("/kurjun/rest/template/download", FileDelete)
		Registered = true
	}
	log.Info(fmt.Sprintf("Configuration before starting: %+v %+v %+v %+v", ConfigurationCDN, ConfigurationDB, ConfigurationNetwork, ConfigurationStorage))
	Server = &http.Server{
		Addr:    ":" + ConfigurationNetwork.Port,
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
		case <-Stop:
			{
				log.Info("Received shut down request. Stopping server...")
				Server.Shutdown(context.Background())
				break loop
			}
		}
	}
	log.Info("Server stopped successfully")
	Stop <- true
}
