package auto

import (
	"github.com/subutai-io/cdn/config"
	"github.com/subutai-io/cdn/db"
	"io/ioutil"
	"log"
	"os"
)

func CleanGarbage() {
	whiteList := []string{"Packages", "Release", "Release.gpg", "Packages.gz"}
	list := db.SearchName("")

	for _, k := range list {
		info := db.Info(k)
		if len(info["Description"]) > 0 {
			whiteList = append(whiteList, info["name"])
		} else {
			whiteList = append(whiteList, info["md5"])
		}
		if info["md5"] == "" {
			whiteList = append(whiteList, info["id"])
		}
	}

	files, err := ioutil.ReadDir(config.Storage.Path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !stringInSlice(file.Name(), whiteList) {
			os.Remove(config.Storage.Path + file.Name())
		}
	}

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
