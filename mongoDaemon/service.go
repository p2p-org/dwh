package mongoDaemon

import (
	"context"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDaemon struct {
	rmqReceiverSender *RMQReceiverSender
	cfg               *DwhQueueServiceConfig
	mongoDB           *mongo.Client
	ctx               context.Context
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

func NewMongoDaemon(configFileName, configPath string) (*MongoDaemon, error) {
	cfg := ReadDwhQueueServiceConfig(configFileName, configPath)

	ctx := context.Background()

	rs, err := NewRMQReceiverSender(cfg)
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

	return &MongoDaemon{
		rmqReceiverSender: rs,
		mongoDB:           mongoDB,
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
	return nil
}
