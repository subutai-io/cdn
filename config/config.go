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
	Path  string
	Quota string
}

type configFile struct {
	DB      dbConfig
	CDN     cdnConfig
	Network networkConfig
	Storage fileConfig
}

const defaultConfig = `
	[db]
	path = /my.db

	[CDN]
	node =

	[network]
	port = 8080

	[storage]
	path = /tmp/
	userquota = 1G
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

	err = gcfg.ReadFileInto(&config, "/etc/gorjun.gcfg")
	log.Check(log.WarnLevel, "Opening Gorjun config file /etc/gorjun.gcfg", err)

	DB = config.DB
	CDN = config.CDN
	// CDN      = "https://cdn.subut.ai:8338"
	Network = config.Network
	Storage = config.Storage
}

func DefaultQuota() int {
	multiplier := 1
	switch config.Storage.Quota[len(config.Storage.Quota)-1:] {
	case "G":
		multiplier = 1024
	case "M":
		multiplier = 1
	case "K":
		multiplier = 1 / 1024
	}
	v, err := strconv.Atoi(config.Storage.Quota[:len(config.Storage.Quota)-1])
	if log.Check(log.WarnLevel, "Converting quota value to int", err) {
		return 1024
	}
	return v * multiplier
}
