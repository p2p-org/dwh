package tokenMetadataService

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	stdLog "log"
	"net/http"
	"reflect"
	"time"

	dwh_common "github.com/corestario/dwh/x/common"
	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenMetadataWorker struct {
	receiver           *dwh_common.RMQReceiver
	client             http.Client
	cfg                *dwh_common.DwhCommonServiceConfig
	mongoClient        *mongo.Client
	mongoCollection    *mongo.Collection
	ctx                context.Context
	erc721SchemaLoader gojsonschema.JSONLoader
	imgSender          *dwh_common.RMQSender
}

func NewTokenMetadataWorker(configFileName, configPath string) (*TokenMetadataWorker, error) {
	cfg := dwh_common.ReadCommonConfig(configFileName, configPath)

	ctx := context.Background()

	receiver, err := dwh_common.NewRMQReceiver(cfg, cfg.UriQueueName, cfg.UriQueueMaxPriority, cfg.UriQueuePrefetchCount)
	if err != nil {
		return nil, fmt.Errorf("could not create rabbitMQ receiver, error: %+v", err)
	}

	imgSender, err := dwh_common.NewRMQSender(cfg, cfg.ImgQueueName, cfg.ImgQueueMaxPriority)
	if err != nil {
		return nil, fmt.Errorf("could not create rabbitMQ sender, error: %+v", err)
	}

	mongoClient, err := dwh_common.GetMongoClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not dial rabbitMQ, error: %+v", err)
	}

	if err = mongoClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("could not connect MongoDB, error: %+v", err)
	}

	mongoCollection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	keys := bson.D{{Key: "dwhData.lastChecked", Value: 1}}
	s, err := mongoCollection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: keys}, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create MongoDB index, error: %+v", err)
	}
	stdLog.Println("created index:", s)

	return &TokenMetadataWorker{
		client:             http.Client{Timeout: time.Second * 15},
		receiver:           receiver,
		mongoCollection:    mongoCollection,
		mongoClient:        mongoClient,
		ctx:                ctx,
		erc721SchemaLoader: gojsonschema.NewStringLoader(erc721Schema),
		imgSender:          imgSender,
		cfg:                cfg,
	}, nil
}

func (tmw *TokenMetadataWorker) Closer() error {
	err := tmw.receiver.Closer()
	if err != nil {
		return fmt.Errorf("could not close rabbitMQ receiver, error: %+v", err)
	}
	if err = tmw.imgSender.Closer(); err != nil {
		return fmt.Errorf("could not close rabbitMQ sender, error: %+v", err)
	}
	if err := tmw.mongoClient.Disconnect(tmw.ctx); err != nil {
		return fmt.Errorf("could not disconnect MongoDB client, error: %+v", err)
	}
	return nil
}

func (tmw *TokenMetadataWorker) Run() error {
	msgs, err := tmw.receiver.GetMessageChan()
	if err != nil {
		return fmt.Errorf("could not get rabbitMQ msg chan, error: %+v", err)
	}

	for d := range msgs {
		err = tmw.processMessage(d.Body, dwh_common.ImgQueuePriority(d.Priority))
		if err != nil {
			stdLog.Println("failed to process rabbitMQ message: ", err)
			// TODO continue needed?
			// continue
		}

		err = d.Ack(false)
		if err != nil {
			stdLog.Println("failed to ack to rabbitMQ: ", err)
			// TODO consider if fatal
			continue
		}

	}
	return nil
}

func (tmw *TokenMetadataWorker) processMessage(msg []byte, priority dwh_common.ImgQueuePriority) error {
	stdLog.Println("got message:", string(msg))
	var (
		rcvd     dwh_common.TaskInfo
		metadata map[string]interface{}
	)

	err := json.Unmarshal(msg, &rcvd)
	if err != nil {
		return fmt.Errorf("unmarshal error: %+v", err)
	}

	metadataBytes, err := tmw.getMetadata(rcvd.URL)
	if err != nil {
		return fmt.Errorf("could not get metadata url, error: %+v", err)
	}

	isValid, err := tmw.isMetadataERC721(metadataBytes)
	if err != nil {
		return fmt.Errorf("could not verify metadata, error: %+v", err)
	}

	if err = bson.UnmarshalExtJSON(metadataBytes, false, &metadata); err != nil {
		return fmt.Errorf("could not unmarshal ext json, error: %+v", err)
	}

	if err := tmw.upsertTokenMetadata(&rcvd, metadata); err != nil {
		return fmt.Errorf("could not upsert token metadata, error: %+v", err)
	}

	if _, ok := metadata["image"]; isValid && ok {
		if err := tmw.imgSender.Publish(metadata["image"].(string), rcvd.Owner, rcvd.TokenID, priority); err != nil {
			return fmt.Errorf("could not publish img task rabbitMQ, error: %+v", err)
		}
	}

	return nil
}

func (tmw *TokenMetadataWorker) getMetadata(url string) ([]byte, error) {
	resp, err := tmw.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get url via http client, error: %+v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body, error: %+v", err)
	}
	return body, nil
}

func (tmw *TokenMetadataWorker) isMetadataERC721(metadata []byte) (bool, error) {
	metadataJsonLoader := gojsonschema.NewBytesLoader(metadata)
	result, err := gojsonschema.Validate(tmw.erc721SchemaLoader, metadataJsonLoader)
	if err != nil {
		return false, fmt.Errorf("could not validate json schema, error: %+v", err)
	}
	return result.Valid(), nil
}

func (tmw *TokenMetadataWorker) upsertTokenMetadata(tokenInfo *dwh_common.TaskInfo, metadata map[string]interface{}) error {
	var (
		oldMetaData map[string]interface{}
		err         error
	)

	filter := map[string]interface{}{"dwhData.tokenID": tokenInfo.TokenID}

	findOpts := []*options.FindOneOptions{{Projection: map[string]interface{}{"dwhData": 0, "_id": 0}}}
	if err = tmw.mongoCollection.FindOne(tmw.ctx, filter, findOpts...).Decode(&oldMetaData); err != nil && err != mongo.ErrNoDocuments {
		return fmt.Errorf("could not find one in mongo collection, error: %+v", err)
	}
	tNow := time.Now().UTC()

	dataForUpsert := make(map[string]interface{})
	if !reflect.DeepEqual(metadata, oldMetaData) {
		dwhData := map[string]interface{}{"tokenID": tokenInfo.TokenID, "owner": tokenInfo.Owner, "url": tokenInfo.URL}
		dwhData["lastUpdated"] = tNow
		dwhData["lastChecked"] = tNow
		metadata["dwhData"] = dwhData
		dataForUpsert = map[string]interface{}{"$set": metadata}
	} else {
		dataForUpsert = map[string]interface{}{"$set": bson.M{"dwhData.lastChecked": tNow}}
	}

	isUpsert := true
	opts := []*options.UpdateOptions{{Upsert: &isUpsert}}

	if _, err = tmw.mongoCollection.UpdateOne(tmw.ctx, filter, dataForUpsert, opts...); err != nil {
		return fmt.Errorf("could not update one in mongo collection, error: %+v", err)
	}

	return nil
}
