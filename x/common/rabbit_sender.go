package dwh_common

import (
	"encoding/json"

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
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
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
		return nil, err
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
		return err
	}
	if err := rs.conn.Close(); err != nil {
		return err
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
		return err
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
	return err
}
