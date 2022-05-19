package repository

import (
	"goodfs/apiserver/global"
	"goodfs/lib/mongodb"
)

var (
	mongo *mongodb.MongoDB
)

func InitMongo(addr string) {
	mongo = mongodb.New(addr)
}

func GetMongo() *mongodb.MongoDB {
	if mongo == nil {
		InitMongo(global.Config.MongoAddress)
	}
	return mongo
}
