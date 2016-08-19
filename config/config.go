package config

import (
	"strconv"

	"github.com/subutai-io/base/agent/log"
	"gopkg.in/gcfg.v1"
)

type cdnConfig struct {
	Node string
}
type networkConfig struct {
	Port string
}
type dbConfig struct {
	Path string
}
type fileConfig struct {
	Path      string
	Userquota string
}

type configFile struct {
	DB      dbConfig
	CDN     cdnConfig
	Network networkConfig
	Storage fileConfig
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
	config configFile

	DB      dbConfig
	CDN     cdnConfig
	Network networkConfig
	Storage fileConfig
)

func init() {
	log.Level(log.InfoLevel)

	err := gcfg.ReadStringInto(&config, defaultConfig)
	log.Check(log.InfoLevel, "Loading default config ", err)

	err = gcfg.ReadFileInto(&config, "/opt/gorjun/etc/gorjun.gcfg")
	log.Check(log.WarnLevel, "Opening Gorjun config file /opt/gorjun/etc/gorjun.gcfg", err)

	DB = config.DB
	CDN = config.CDN
	// CDN      = "https://cdn.subut.ai:8338"
	Network = config.Network
	Storage = config.Storage
}

func DefaultQuota() int {
	multiplier := 1
	switch config.Storage.Userquota[len(config.Storage.Userquota)-1:] {
	case "G":
		multiplier = 1073741824
	case "M":
		multiplier = 1048576
	case "K":
		multiplier = 1024
	}
	v, err := strconv.Atoi(config.Storage.Userquota[:len(config.Storage.Userquota)-1])
	if log.Check(log.WarnLevel, "Converting quota value to int", err) {
		return 1073741824
	}
	return v * multiplier
}
