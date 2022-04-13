package repository

import (
	"goodfs/apiserver/config"
	"goodfs/lib/mongodb"
)

var (
	mongo *mongodb.MongoDB
)

func GetMongo() *mongodb.MongoDB {
	if mongo == nil {
		mongo = mongodb.New(config.MongoAddress)
	}
	return mongo
}
