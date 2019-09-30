package imgservice

import (
	"github.com/streadway/amqp"
)

type RMQReceiver struct {
	config *DwhImgServiceConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	imgQ   *amqp.Queue
}

func NewRMQReceiver(cfg *DwhImgServiceConfig) (*RMQReceiver, error) {
	addr := QueueAddrStringFromConfig(cfg)

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	qArgs := map[string]interface{}{"x-max-priority": cfg.ImgQueueMaxPriority}

	q, err := ch.QueueDeclare(
		cfg.ImgQueueName,
		true,
		false,
		false,
		false,
		qArgs,
	)
	if err != nil {
		return nil, err
	}

	err = ch.Qos(
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
		ch:     ch,
		imgQ:   &q,
	}, nil

}

func (rs *RMQReceiver) Closer() error {
	if err := rs.ch.Close(); err != nil {
		return err
	}

	if err := rs.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (rs *RMQReceiver) GetMessageChan() (<-chan amqp.Delivery, error) {
	msgs, err := rs.ch.Consume(
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
