package dwh_common

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMongoClient(cfg *DwhCommonServiceConfig) (*mongo.Client, error) {
	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		cfg.MongoUserName,
		cfg.MongoUserPass,
		cfg.MongoHost,
		cfg.MongoDatabase,
	)
	opt := options.Client().ApplyURI(uri)
	creds := options.Credential{
		Username: cfg.MongoUserName,
		Password: cfg.MongoUserPass,
	}
	opt = opt.SetAuth(creds)
	client, err := mongo.NewClient(opt)
	if err != nil {
		return nil, fmt.Errorf("could not create mongo client, error: %+v", err)
	}
	return client, nil
}
