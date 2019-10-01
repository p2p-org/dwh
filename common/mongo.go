package common

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoUserName   = "dgaming"
	MongoUserPass   = "dgaming"
	MongoHost       = "localhost:27017"
	MongoDatabase   = "dgaming"
	MongoCollection = "metadata"
)

func GetMongoDB() (*mongo.Client, error) {
	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		MongoUserName,
		MongoUserPass,
		MongoHost,
		MongoDatabase,
	)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	return client, err
}
