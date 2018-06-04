package main

import (
	"github.com/subutai-io/cdn/db"
	"io/ioutil"
	"github.com/subutai-io/cdn/config"
	"os"
	"github.com/subutai-io/agent/log"
	"fmt"
)

func CleanGarbage() {
	whiteList := []string{"Packages", "Release", "Release.gpg", "Packages.gz"}
	list := db.SearchName("")
	for _, k := range list {
		info := db.Info(k)
		whiteList = append(whiteList, info["name"])
		whiteList = append(whiteList, info["md5"])
		whiteList = append(whiteList, info["id"])
	}
	files, _ := ioutil.ReadDir(config.Storage.Path)
	for _, file := range files {
		if !stringInSlice(file.Name(), whiteList) {
			os.Remove(config.Storage.Path + file.Name())
		}
	}
	for _, k := range list {
		md5, _ := db.Hash(k)
		name := db.FileField(k, "name")
		ok := false
		for _, file := range files {
			if file.Name() == md5 || (len(name) > 0 && file.Name() == name[0]) {
				ok = true
				break
			}
		}
		if !ok {
			log.Info(fmt.Sprintf("Deleted file %s", k))
			owner := db.FileField(k, "owner")
			if len(owner) > 0 {
				db.Delete(db.FileField(k, "owner")[0], db.CheckRepoOfHash(k), k)
			} else {
				db.Delete("", db.CheckRepoOfHash(k), k)
			}
		}
	}
	db.CleanSearchIndex()
	db.CleanUserFiles()
	db.CleanTokens()
	db.CleanAuthID()
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	CleanGarbage()
}
