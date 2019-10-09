package dwh_common

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RMQReceiver struct {
	config *DwhCommonServiceConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	queue  *amqp.Queue
}

func NewRMQReceiver(cfg *DwhCommonServiceConfig, queueName string, queueMaxPriority, queuePrefetchCount int) (*RMQReceiver, error) {
	addr := QueueAddrStringFromConfig(cfg)

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("could not dial rabbitMQ, error: %+v", err)
	}

	uriCh, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not create rabbitMQ channel error: %+v", err)
	}

	uriQArgs := map[string]interface{}{"x-max-priority": queueMaxPriority}

	uriQ, err := uriCh.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		uriQArgs,
	)
	if err != nil {
		return nil, fmt.Errorf("could not declare rabbitMQ queue, error: %+v", err)
	}

	err = uriCh.Qos(
		queuePrefetchCount,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("could not set rabbitMQ Qos, error: %+v", err)
	}

	return &RMQReceiver{
		config: cfg,
		conn:   conn,
		ch:     uriCh,
		queue:  &uriQ,
	}, nil

}

func (rs *RMQReceiver) Closer() error {
	if err := rs.ch.Close(); err != nil {
		return fmt.Errorf("could not close rabbitMQ channel, error: %+v", err)
	}

	if err := rs.conn.Close(); err != nil {
		return fmt.Errorf("could not close rabbitMQ connection, error: %+v", err)
	}
	return nil
}

func (rs *RMQReceiver) GetMessageChan() (<-chan amqp.Delivery, error) {
	msgs, err := rs.ch.Consume(
		rs.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get rabbitMQ msg chan, error: %+v", err)
	}
	return msgs, nil
}
