package lib

import (
	"net/http"
	"strings"
)

func Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		return
	}
	repo := strings.Split(r.RequestURI, "/")[3]
	if !In(repo, allRepos) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect repo"))
		return
	}
	if info := GetInfo(repo, r); len(info) > 0 {
		w.Write(info)
		return
	}
	w.Write([]byte("404 Not found"))
}

func List(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect method"))
		return
	}
	repo := strings.Split(r.RequestURI, "/")[3]
	if !In(repo, allRepos) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect repo"))
		return
	}
	w.Write(GetList(repo, r))
}
