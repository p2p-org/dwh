package mongoDaemon

import (
	"context"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	stdLog "log"
	"time"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDaemon struct {
	rmqReceiverSender *RMQReceiverSender
	cfg               *dwh_common.DwhCommonServiceConfig
	mongoClient       *mongo.Client
	mongoDB           *mongo.Database
	mongoCollection   *mongo.Collection
	ctx               context.Context
}

func NewMongoDaemon(configFileName, configPath string) (*MongoDaemon, error) {
	cfg := dwh_common.ReadCommonConfig(configFileName, configPath)

	ctx := context.Background()

	rs, err := NewRMQReceiverSender(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create mongo daemon, error: %+v", err)
	}

	mongoClient, err := dwh_common.GetMongoClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create mongo client, error: %+v", err)
	}
	if err := mongoClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("could not connect to mongo, error: %+v", err)
	}

	mongoDB := mongoClient.Database(cfg.MongoDatabase)
	collection := mongoDB.Collection(cfg.MongoCollection)
	return &MongoDaemon{
		rmqReceiverSender: rs,
		mongoClient:       mongoClient,
		mongoDB:           mongoDB,
		mongoCollection:   collection,
		ctx:               ctx,
		cfg:               cfg,
	}, nil

}

func (md *MongoDaemon) Closer() error {
	err := md.rmqReceiverSender.Closer()
	if err != nil {
		return fmt.Errorf("could not close receiver_sender, error: %+v", err)
	}

	err = md.mongoClient.Disconnect(md.ctx)
	if err != nil {
		return fmt.Errorf("could not disconnect from mongo, error: %+v", err)
	}
	return nil
}

func (md *MongoDaemon) Run() error {
	msgs, err := md.rmqReceiverSender.GetTaskMessageChan()
	if err != nil {
		return fmt.Errorf("could not get rabbitMQ chan, error: %+v", err)
	}

	for d := range msgs {
		err = md.processMessage(d.Body)
		if err != nil {
			stdLog.Println("failed to process rabbitMQ message: ", err)
			// TODO
			// continue
		}

		if err := d.Ack(false); err != nil {
			stdLog.Println("failed to ack to rabbitMQ: ", err)
			// TODO
			// continue
		}

		if err := md.rmqReceiverSender.publishDelayed(); err != nil {
			stdLog.Println("failed to publish delayed message to rabbitMQ: ", err)
		}
	}
	return nil
}

func (md *MongoDaemon) processMessage(msg []byte) error {
	overallCount, err := md.mongoCollection.EstimatedDocumentCount(md.ctx)
	count := overallCount * md.cfg.DaemonUpdatePercent / 100

	findOpts := []*options.FindOptions{{Limit: &count}, {Sort: bson.D{{"dwhData.lastChecked", 1}}}}
	cur, err := md.mongoCollection.Find(md.ctx, bson.D{}, findOpts...)
	if err != nil {
		return fmt.Errorf("could not find in mongo collection, error: %+v", err)
	}

	defer cur.Close(md.ctx)
	tNow := time.Now().UTC()
	for cur.Next(md.ctx) {
		var oldMetaData map[string]interface{}
		err := cur.Decode(&oldMetaData)
		if err != nil {
			return fmt.Errorf("could not decode old metadata, error: %+v", err)
		}
		id, ok := oldMetaData["_id"]
		if !ok {
			err = fmt.Errorf("no _id field in document")
			stdLog.Println(err)
			continue
		}

		dwhData, ok := oldMetaData["dwhData"]
		if !ok {
			err = fmt.Errorf("no dwhData field in document")
			stdLog.Println(err)
			continue
		}
		dwh, ok := dwhData.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("invalid dwhData field in document")
			stdLog.Println(err)
			continue
		}
		owner, ok := dwh["owner"]
		if !ok {
			err = fmt.Errorf("no owner field in dwhData")
			stdLog.Println(err)
			continue
		}
		uri, ok := dwh["url"]
		if !ok {
			err = fmt.Errorf("no url field in dwhData")
			stdLog.Println(err)
			continue
		}
		tokenId, ok := dwh["tokenID"]
		if !ok {
			err = fmt.Errorf("no tokenID field in dwhData")
			stdLog.Println(err)
			continue
		}

		filter := map[string]interface{}{"_id": id}
		dataForUpdate := map[string]interface{}{"$set": bson.M{"dwhData.lastChecked": tNow}}

		res, err := md.mongoCollection.UpdateOne(md.ctx, filter, dataForUpdate)
		if err != nil {
			return fmt.Errorf("failed to update in collection, error: %v", err)
		}
		if res.MatchedCount == 0 || res.MatchedCount != res.ModifiedCount {
			return fmt.Errorf("failed to update: no matches")
		}
		if err := md.rmqReceiverSender.PublishUriTask(
			fmt.Sprintf("%v", uri),
			fmt.Sprintf("%v", owner),
			fmt.Sprintf("%v", tokenId),
		); err != nil {
			return fmt.Errorf("failed to publish uri tasks, error: %v", err)
		}
	}
	if err := cur.Err(); err != nil {
		return fmt.Errorf("mongodb cursor error: %v", err)
	}
	return nil
}
