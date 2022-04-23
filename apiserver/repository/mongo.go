package repository

import (
	"goodfs/apiserver/global"
	"goodfs/lib/mongodb"
)

var (
	mongo *mongodb.MongoDB
)

func GetMongo() *mongodb.MongoDB {
	if mongo == nil {
		mongo = mongodb.New(global.Config.MongoAddress)
	}
	return mongo
}
