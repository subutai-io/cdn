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
	db.RegisterUser([]byte(Lorem.Username), []byte(LoremKey))
	db.RegisterUser([]byte(Ipsum.Username), []byte(IpsumKey))
	Subutai.Token = "15a5237ee5314282e52156cfad72e86b53ef0ad47baecc31233dbb1c06f4327c"
	Lorem.Token = "7327753f12b67d440f481c27e461925513d30cf2d56b6ac16060aad1021c293d"
	Ipsum.Token = "37fc4c3ef862c079ea44da0b7863948d10e8493c514d83e751a6253363f564cf"
	db.SaveToken(Subutai.Username, Subutai.Token)
	db.SaveToken(Lorem.Username, Lorem.Token)
	db.SaveToken(Ipsum.Username, Ipsum.Token)
}

func Clean() {
	log.Info("Cleaning working directories")
	os.RemoveAll("/tmp/cdn-test-data/public/lorem/")
	os.RemoveAll("/tmp/cdn-test-data/public/ipsum/")
	os.RemoveAll("/tmp/cdn-test-data/public/subutai/")
	os.RemoveAll("/tmp/cdn-test-data/public/")
	os.RemoveAll("/tmp/cdn-test-data/private/lorem/")
	os.RemoveAll("/tmp/cdn-test-data/private/ipsum/")
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
	os.MkdirAll("/tmp/cdn-test-data/public/lorem/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/public/ipsum/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/public/subutai/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/lorem/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/ipsum/", 0755)
	os.MkdirAll("/tmp/cdn-test-data/private/subutai/", 0755)
	config.ConfigurationDB.Path = "/tmp/cdn-test-data/db/my.db"
	config.ConfigurationNetwork.Port = "8080"
	config.ConfigurationStorage.Path = "/tmp/cdn-test-data/files/"
	config.ConfigurationStorage.Userquota = "2G"
	db.DB = db.InitDB()
	log.Info(fmt.Sprintf("Testing environment and configuration are set up: %+v %+v %+v", config.ConfigurationDB, config.ConfigurationNetwork, config.ConfigurationStorage))
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
