package tokenMetadataSaverService

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgamingfoundation/dwh/common"
	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"time"
)

type TokenMetadataWorker struct {
	receiver           *RMQReceiver
	client             http.Client
	cfg                *DwhQueueServiceConfig
	mongoDB            *mongo.Client
	ctx                context.Context
	erc721SchemaLoader gojsonschema.JSONLoader
	imgSender          *RMQSender
}

func getMongoDB(cfg *DwhQueueServiceConfig) (*mongo.Client, error) {
	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		cfg.MongoUserName,
		cfg.MongoUserPass,
		cfg.MongoHost,
		cfg.MongoDatabase,
	)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	return client, err
}

func NewTokenMetadataWorker(configFileName, configPath string) (*TokenMetadataWorker, error) {
	cfg := ReadDwhTokenMetadataServiceConfig(configFileName, configPath)

	ctx := context.Background()

	receiver, err := NewRMQReceiver(cfg)
	if err != nil {
		return nil, err
	}

	sender, err := NewRMQSender(configFileName, configPath)
	if err != nil {
		return nil, err
	}

	mongoDB, err := getMongoDB(cfg)
	if err != nil {
		return nil, err
	}
	if err := mongoDB.Connect(ctx); err != nil {
		return nil, err
	}

	return &TokenMetadataWorker{
		client:             http.Client{Timeout: time.Second * 15},
		receiver:           receiver,
		mongoDB:            mongoDB,
		ctx:                ctx,
		erc721SchemaLoader: gojsonschema.NewStringLoader(erc721Schema),
		imgSender:          sender,
	}, nil
}

func (tmw *TokenMetadataWorker) Closer() error {
	err := tmw.receiver.Closer()
	if err != nil {
		return err
	}
	if err = tmw.imgSender.Closer(); err != nil {
		return err
	}
	return nil
}

func (tmw *TokenMetadataWorker) Run() error {
	msgs, err := tmw.receiver.GetUriMessageChan()
	if err != nil {
		return err
	}

	for d := range msgs {
		err = tmw.processMessage(d.Body)
		if err != nil {
			fmt.Println("failed to process rabbitMQ message: ", err)
			continue
		}

		err = d.Ack(false)
		if err != nil {
			fmt.Println("failed to ack to rabbitMQ: ", err)
			continue
		}

	}
	return nil
}

func (tmw *TokenMetadataWorker) processMessage(msg []byte) error {
	var (
		rcvd     TokenInfo
		metadata map[string]interface{}
	)

	err := json.Unmarshal(msg, &rcvd)
	if err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}

	metadataBytes, err := tmw.getMetadata(rcvd.URL)
	if err != nil {
		return err
	}

	isValid, err := tmw.isMetadataERC721(metadataBytes)
	if err != nil {
		return err
	}

	if err := tmw.insertTokenMetadata(tmw.ctx, tmw.mongoDB, rcvd.TokenID, metadataBytes); err != nil {
		return err
	}

	if isValid {
		if err := tmw.imgSender.Publish(metadata["image"].(string), rcvd.Owner, tmw.cfg.ImgQueueMaxPriority); err != nil {
			return err
		}
	}

	return nil
}

func (tmw *TokenMetadataWorker) getMetadata(url string) ([]byte, error) {
	resp, err := tmw.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (tmw *TokenMetadataWorker) isMetadataERC721(metadata []byte) (bool, error) {
	metadataJsonLoader := gojsonschema.NewBytesLoader(metadata)
	result, err := gojsonschema.Validate(tmw.erc721SchemaLoader, metadataJsonLoader)
	if err != nil {
		return false, err
	}
	return result.Valid(), nil
}

func (tmw *TokenMetadataWorker) insertTokenMetadata(ctx context.Context, mongo *mongo.Client, tokenID string, metadata []byte) error {
	var dataForInsert map[string]interface{}
	if err := json.Unmarshal(metadata, &dataForInsert); err != nil {
		return err
	}
	dataForInsert["tokenID"] = tokenID
	result, err := mongo.Database(common.MongoDatabase).Collection(common.MongoCollection).InsertOne(ctx, dataForInsert)
	if err != nil {
		return err
	}
	fmt.Println(result.InsertedID)
	return nil
}
