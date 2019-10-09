package mongoDaemon

import (
	"context"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
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
		return nil, err
	}

	mongoClient, err := dwh_common.GetMongoClient(cfg)
	if err != nil {
		return nil, err
	}
	if err := mongoClient.Connect(ctx); err != nil {
		return nil, err
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
		return err
	}
	return nil
}

func (md *MongoDaemon) Run() error {
	msgs, err := md.rmqReceiverSender.GetTaskMessageChan()
	if err != nil {
		return err
	}

	for d := range msgs {
		err = md.processMessage(d.Body)
		if err != nil {
			fmt.Println("failed to process rabbitMQ message: ", err)
			continue
		}

		if err := d.Ack(false); err != nil {
			fmt.Println("failed to ack to rabbitMQ: ", err)
			continue
		}

		if err := md.rmqReceiverSender.publishDelayed(); err != nil {
			fmt.Println("failed to publish delayed message to rabbitMQ: ", err)
			continue
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
		return err
	}
	defer cur.Close(md.ctx)
	tNow := time.Now().UTC()
	for cur.Next(md.ctx) {
		var oldMetaData map[string]interface{}
		err := cur.Decode(&oldMetaData)
		if err != nil {
			return err
		}
		id, ok := oldMetaData["_id"]
		if !ok {
			err = fmt.Errorf("no _id field in document")
			log.Println(err)
			continue
		}

		dwhData, ok := oldMetaData["dwhData"]
		if !ok {
			err = fmt.Errorf("no dwhData field in document")
			log.Println(err)
			continue
		}
		dwh, ok := dwhData.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("invalid dwhData field in document")
			log.Println(err)
			continue
		}
		owner, ok := dwh["owner"]
		if !ok {
			err = fmt.Errorf("no owner field in dwhData")
			log.Println(err)
			continue
		}
		uri, ok := dwh["url"]
		if !ok {
			err = fmt.Errorf("no url field in dwhData")
			log.Println(err)
			continue
		}
		tokenId, ok := dwh["tokenID"]
		if !ok {
			err = fmt.Errorf("no tokenID field in dwhData")
			log.Println(err)
			continue
		}

		filter := map[string]interface{}{"_id": id}
		dataForUpdate := map[string]interface{}{"$set": bson.M{"dwhData.lastChecked": tNow}}

		res, err := md.mongoCollection.UpdateOne(md.ctx, filter, dataForUpdate)
		if err != nil {
			return err
		}
		if res.MatchedCount == 0 || res.MatchedCount != res.ModifiedCount {
			return fmt.Errorf("failed to update: no matches")
		}
		if err := md.rmqReceiverSender.PublishUriTask(
			fmt.Sprintf("%v", uri),
			fmt.Sprintf("%v", owner),
			fmt.Sprintf("%v", tokenId),
		); err != nil {
			return err
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}
