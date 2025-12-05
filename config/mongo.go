package config

import (
	"fmt"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo(ctx context.Context, DBCollection... string) *mongo.Database  {
	connection := fmt.Sprintf("mongodb://%s:%s", MONGOHost, MONGOPort)
	fmt.Println("Connection Mongo:", connection)
	clientOptions := options.Client()
	clientOptions.ApplyURI(connection)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil
	}

	col := MONGODB
	if len(DBCollection) > 0 {
		col = DBCollection[0]
	}

	return client.Database(col)
}