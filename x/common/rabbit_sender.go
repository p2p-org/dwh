package dwh_common

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

type RMQSender struct {
	config *DwhCommonServiceConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	imgQ   *amqp.Queue
}

func NewRMQSender(cfg *DwhCommonServiceConfig, queueName string, queueMaxPriority int) (*RMQSender, error) {
	u := QueueAddrStringFromConfig(cfg)

	conn, err := amqp.Dial(u)
	if err != nil {
		return nil, fmt.Errorf("could not dial rabbitMQ, error: %+v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not create rabbitMQ channel, error: %+v", err)
	}

	qArgs := map[string]interface{}{"x-max-priority": queueMaxPriority}
	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		qArgs,
	)
	if err != nil {
		return nil, fmt.Errorf("could not declare rabbitMQ queue, error: %+v", err)
	}

	return &RMQSender{
		config: cfg,
		conn:   conn,
		ch:     ch,
		imgQ:   &q,
	}, nil

}

func (rs *RMQSender) Closer() error {
	if err := rs.ch.Close(); err != nil {
		return fmt.Errorf("could not close rabbitMQ channel, error: %+v", err)
	}
	if err := rs.conn.Close(); err != nil {
		return fmt.Errorf("could not close rabbitMQ connection, error: %+v", err)
	}
	return nil
}

func (rs *RMQSender) Publish(taskUrl, owner, tokenId string, priority ImgQueuePriority) error {
	ba, err := json.Marshal(&TaskInfo{
		Owner:   owner,
		URL:     taskUrl,
		TokenID: tokenId,
	})
	if err != nil {
		return fmt.Errorf("could not marshal rabbitMQ task, error: %+v", err)
	}

	err = rs.ch.Publish(
		"",
		rs.imgQ.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         ba,
			Priority:     uint8(priority),
		})
	return fmt.Errorf("could not publish rabbitMQ message, error: %+v", err)
}
