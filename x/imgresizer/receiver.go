package imgresizer

import (
	dwh_common "github.com/dgamingfoundation/dwh/x/common"
	"github.com/streadway/amqp"
)

type RMQReceiver struct {
	config *dwh_common.DwhCommonServiceConfig
	conn   *amqp.Connection
	imgCh  *amqp.Channel
	imgQ   *amqp.Queue
}

func NewRMQReceiver(cfg *dwh_common.DwhCommonServiceConfig) (*RMQReceiver, error) {
	addr := dwh_common.QueueAddrStringFromConfig(cfg)

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	imgCh, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	imgQArgs := map[string]interface{}{"x-max-priority": cfg.ImgQueueMaxPriority}

	imgQ, err := imgCh.QueueDeclare(
		cfg.ImgQueueName,
		true,
		false,
		false,
		false,
		imgQArgs,
	)
	if err != nil {
		return nil, err
	}

	err = imgCh.Qos(
		cfg.ImgQueuePrefetchCount,
		0,
		false,
	)
	if err != nil {
		return nil, err
	}

	return &RMQReceiver{
		config: cfg,
		conn:   conn,
		imgCh:  imgCh,
		imgQ:   &imgQ,
	}, nil

}

func (rs *RMQReceiver) Closer() error {
	if err := rs.imgCh.Close(); err != nil {
		return err
	}

	if err := rs.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (rs *RMQReceiver) GetImgMessageChan() (<-chan amqp.Delivery, error) {
	msgs, err := rs.imgCh.Consume(
		rs.imgQ.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	return msgs, err
}
