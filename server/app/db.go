package app

import (
	"time"

	"github.com/asdine/storm/q"
	"github.com/subutai-io/agent/log"
)

type File struct {
	FileID        string    `storm:"unique"`
	Name          string    `storm:"index"`
	Owner         string    `storm:"index"`
	Version       string    `storm:"index"`
	Md5           string    `storm:"index"`
	Sha256        string    `storm:"index"`
	Tags          []string  `storm:"index"`
	Date          time.Time `storm:"index"`
	Timestamp     string    `storm:"index"`
	Size          int64     `storm:"index"`
	Description   string    `storm:"index"`
	Architecture  string    `storm:"index"`
	Parent        string    `storm:"index"`
	ParentVersion string    `storm:"index"`
	ParentOwner   string    `storm:"index"`
	PrefSize      string    `storm:"index"`
}

type User struct {
	UserName string   `storm:"unique"`
	Files    []File   `storm:"index"`
	Tokens   []string `storm:"index"`
	Keys     []string `storm:"index"`
	AuthIDs  []string `storm:"index"`
}

type Parameter struct {
	Field string
	Value interface{}
}

func PrepareQuery(parameters ...Parameter) q.Matcher {
	log.Info("Started PrepareQuery")
	var query q.Matcher
	for i := range parameters {
		parameter := parameters[i]
		if query == nil {
			query = q.Eq(parameter.Field, parameter.Value)
		} else {
			query = q.And(query, q.Eq(parameter.Field, parameter.Value))
		}
	}
	log.Info("Finished PrepareQuery")
	return query
}

func GetFileInfo(query q.Matcher) (files []File) {
	log.Info("Started GetFileInfo")
	db.Select(query).Find(&files)
	log.Info("Finished GetFileInfo")
	return files
}

func GetUserInfo(query q.Matcher) (users []User) {
	log.Info("Started GetUserInfo")
	DB.Select(query).Find(&users)
	log.Info("Finished GetUserInfo")
	return users
}
