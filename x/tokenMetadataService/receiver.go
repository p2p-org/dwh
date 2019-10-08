package tokenMetadataService

import (
	dwh_common "github.com/dgamingfoundation/dwh/x/common"
	"github.com/streadway/amqp"
)

type RMQReceiver struct {
	config *dwh_common.DwhCommonServiceConfig
	conn   *amqp.Connection
	imgCh  *amqp.Channel
	imgQ   *amqp.Queue
	uriCh  *amqp.Channel
	uriQ   *amqp.Queue
}

func NewRMQReceiver(cfg *dwh_common.DwhCommonServiceConfig) (*RMQReceiver, error) {
	addr := dwh_common.QueueAddrStringFromConfig(cfg)

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	uriCh, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	uriQArgs := map[string]interface{}{"x-max-priority": cfg.UriQueueMaxPriority}

	uriQ, err := uriCh.QueueDeclare(
		cfg.UriQueueName,
		true,
		false,
		false,
		false,
		uriQArgs,
	)
	if err != nil {
		return nil, err
	}

	err = uriCh.Qos(
		cfg.UriQueuePrefetchCount,
		0,
		false,
	)
	if err != nil {
		return nil, err
	}

	return &RMQReceiver{
		config: cfg,
		conn:   conn,
		uriCh:  uriCh,
		uriQ:   &uriQ,
	}, nil

}

func (rs *RMQReceiver) Closer() error {
	if err := rs.imgCh.Close(); err != nil {
		return err
	}

	if err := rs.uriCh.Close(); err != nil {
		return err
	}

	if err := rs.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (rs *RMQReceiver) GetUriMessageChan() (<-chan amqp.Delivery, error) {
	msgs, err := rs.uriCh.Consume(
		rs.uriQ.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	return msgs, err
}
