package app

import (
	"os"
	"fmt"
	"github.com/subutai-io/cdn/db"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/config"
)

func PrepareUsersAndTokens() {
	db.RegisterUser([]byte(Subutai.Username), []byte(SubutaiKey))
	db.RegisterUser([]byte(Akenzhaliev.Username), []byte(AkenzhalievKey))
	db.RegisterUser([]byte(Abaytulakova.Username), []byte(AbaytulakovaKey))
	Subutai.Token = "15a5237ee5314282e52156cfad72e86b53ef0ad47baecc31233dbb1c06f4327c"
	Akenzhaliev.Token = "7327753f12b67d440f481c27e461925513d30cf2d56b6ac16060aad1021c293d"
	Abaytulakova.Token = "37fc4c3ef862c079ea44da0b7863948d10e8493c514d83e751a6253363f564cf"
	db.SaveToken(Subutai.Username, Subutai.Token)
	db.SaveToken(Akenzhaliev.Username, Akenzhaliev.Token)
	db.SaveToken(Abaytulakova.Username, Abaytulakova.Token)
}

func Clean() {
	log.Info("Cleaning working directories")
	os.RemoveAll("/tmp/cdn-test-data/public/akenzhaliev/")
	os.RemoveAll("/tmp/cdn-test-data/public/abaytulakova/")
	os.RemoveAll("/tmp/cdn-test-data/public/subutai/")
	os.RemoveAll("/tmp/cdn-test-data/public/")
	os.RemoveAll("/tmp/cdn-test-data/private/akenzhaliev/")
	os.RemoveAll("/tmp/cdn-test-data/private/abaytulakova/")
	os.RemoveAll("/tmp/cdn-test-data/private/subutai/")
	os.RemoveAll("/tmp/cdn-test-data/private/")
	os.RemoveAll("/tmp/cdn-test-data/files/")
	os.RemoveAll("/tmp/cdn-test-data/db/")
	os.RemoveAll("/tmp/cdn-test-data/")
}

func SetUp() {
	log.Level(log.DebugLevel)
	log.Info("Setting up testing environment and configuration")
	Clean()
	InitFilters()
	InitPreUploaders()
	os.MkdirAll("/tmp/cdn-test-data/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/db/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/files/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/public/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/public/akenzhaliev/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/public/abaytulakova/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/public/subutai/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/akenzhaliev/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/abaytulakova/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/subutai/", 0755)
	config.DB.Path = "/tmp/cdn-test-data/db/my.db"
	config.Network.Port = "8080"
	config.Storage.Path = "/tmp/cdn-test-data/files/"
	config.Storage.Userquota = "2G"
	db.DB = db.InitDB()
	log.Info(fmt.Sprintf("Testing environment and configuration are set up: %+v %+v %+v", config.DB, config.Network, config.Storage))
	if Integration == 1 {
		RunServer()
	}
}

func TearDown() {
	if Integration == 1 {
		for {
			if Stop != nil {
				Stop <- true
				break
			}
		}
		<-Stop
		close(Stop)
	}
	log.Info("Destroying testing environment")
	Clean()
	log.Info("Testing environment destroyed")
}

