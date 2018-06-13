package app

import (
	"github.com/subutai-io/agent/log"
	"github.com/asdine/storm/q"
)

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

func GetUserInfo(query q.Matcher) (users []User) {
	log.Info("Started GetUserInfo")
	DB.Select(query).Find(&users)
	log.Info("Finished GetUserInfo")
	return users
}
