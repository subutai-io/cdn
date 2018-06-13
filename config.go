package app

import (
	"strconv"

	"github.com/subutai-io/agent/log"
	"gopkg.in/gcfg.v1"
	"fmt"
	"github.com/asdine/storm"
	"os"
)

type cdnConfiguration struct {
	Node string
}
type networkConfiguration struct {
	Port string
}
type dbConfiguration struct {
	Path string
}
type fileConfiguration struct {
	Path      string
	Userquota string
}

type configFile struct {
	DB      dbConfiguration
	CDN     cdnConfiguration
	Network networkConfiguration
	Storage fileConfiguration
}

const defaultConfig = `
	[db]
	path = /opt/gorjun/data/db/my.db

	[CDN]
	node =

	[network]
	port = 8080

	[storage]
	path = /opt/gorjun/data/files/
	userquota = 2G
`

var (
	ConfigurationFile configFile

	ConfigurationDB      dbConfiguration
	ConfigurationCDN     cdnConfiguration
	ConfigurationNetwork networkConfiguration
	ConfigurationStorage fileConfiguration
)

func InitDB() {
	var err error
	DB, err = storm.Open(ConfigurationDB.Path)
	if err != nil {
		log.Panic("Couldn't open DB")
		os.Exit(1)
	}
}

func InitConfig() {
	log.Info("Initialization started")
	err := gcfg.ReadStringInto(&ConfigurationFile, defaultConfig)
	log.Check(log.InfoLevel, "Loading default config ", err)
	err = gcfg.ReadFileInto(&ConfigurationFile, "/opt/gorjun/etc/gorjun.gcfg")
	log.Check(log.WarnLevel, "Opening Gorjun config file /opt/gorjun/etc/gorjun.gcfg", err)
	ConfigurationDB = ConfigurationFile.DB
	ConfigurationCDN = ConfigurationFile.CDN
	ConfigurationNetwork = ConfigurationFile.Network
	ConfigurationStorage = ConfigurationFile.Storage
	log.Info(fmt.Sprintf("Initialization completed: %s %s %s %s %s", ConfigurationDB.Path, ConfigurationCDN.Node, ConfigurationNetwork.Port, ConfigurationStorage.Path, ConfigurationStorage.Userquota))
}

func DefaultQuota() int {
	multiplier := 1
	switch ConfigurationStorage.Userquota[len(ConfigurationStorage.Userquota)-1:] {
	case "G":
		multiplier = 1073741824
	case "M":
		multiplier = 1048576
	case "K":
		multiplier = 1024
	}
	v, err := strconv.Atoi(ConfigurationStorage.Userquota[:len(ConfigurationStorage.Userquota)-1])
	if log.Check(log.WarnLevel, "Converting quota value to int", err) {
		return 1073741824
	}
	return v * multiplier
}
