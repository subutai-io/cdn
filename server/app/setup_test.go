package app

import (
	"os"
	"fmt"
	"github.com/subutai-io/agent/log"
)

func PrepareUsersAndTokens() {
	DB.RegisterUser([]byte(Subutai.Username), []byte(SubutaiKey))
	DB.RegisterUser([]byte(Lorem.Username), []byte(LoremKey))
	DB.RegisterUser([]byte(Ipsum.Username), []byte(IpsumKey))
	Subutai.Token = "15a5237ee5314282e52156cfad72e86b53ef0ad47baecc31233dbb1c06f4327c"
	Lorem.Token = "7327753f12b67d440f481c27e461925513d30cf2d56b6ac16060aad1021c293d"
	Ipsum.Token = "37fc4c3ef862c079ea44da0b7863948d10e8493c514d83e751a6253363f564cf"
	DB.SaveToken(Subutai.Username, Subutai.Token)
	DB.SaveToken(Lorem.Username, Lorem.Token)
	DB.SaveToken(Ipsum.Username, Ipsum.Token)
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
	ConfigurationDB.Path = "/tmp/cdn-test-data/db/my.db"
	ConfigurationNetwork.Port = "8080"
	ConfigurationStorage.Path = "/tmp/cdn-test-data/files/"
	ConfigurationStorage.Userquota = "2G"
	log.Info(fmt.Sprintf("Testing environment and configuration are set up: %+v %+v %+v", ConfigurationDB, ConfigurationNetwork, ConfigurationStorage))
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
