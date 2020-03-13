package models

import (
	"context"
	"weqi_service/util"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Database

func Database(connString, dbname string) {
	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(connString)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		util.Log().Panic("连接mongodb不成功", err)
		return
	}
	db := client.Database(dbname)
	Client = db
	util.Log().Info(db.Name())
}
