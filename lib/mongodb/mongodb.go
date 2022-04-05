package mongodb

import (
	"context"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	client    *mongo.Client
	db        *mongo.Database
	context   context.Context
	connected bool
}

func New(addr string) *MongoDB {
	parms := strings.Split(addr, "#")

	if len(parms) < 5 {
		log.Panic("Init MongoDB fail, require addr, dbName, authType, username and password join by '#'")
	}

	rootCxt := context.Background()
	ctx, cancel := context.WithTimeout(rootCxt, 5*time.Second)
	defer cancel()

	credential := options.Credential{
		AuthMechanism: parms[2],
		Username:      parms[3],
		Password:      parms[4],
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(parms[0]).SetAuth(credential))

	if err != nil {
		panic(err)
	}

	return &MongoDB{
		client:    client,
		db:        client.Database(parms[1]),
		context:   rootCxt,
		connected: false,
	}
}

func (m *MongoDB) Check() bool {
	ctx, cancel := context.WithTimeout(m.context, 1*time.Second)
	defer cancel()
	err := m.client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Println(err)
		return false
	}
	m.connected = true
	return true
}

func (m *MongoDB) Close() {
	<-m.context.Done()
	ctx, cancel := context.WithTimeout(m.context, 5*time.Second)
	defer cancel()
	if err := m.client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func (m *MongoDB) Collection(name string) *mongo.Collection {

	return m.db.Collection(name)
}
