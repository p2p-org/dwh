package mongoDaemon

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type RMQReceiverSender struct {
	config   *DwhQueueServiceConfig
	conn     *amqp.Connection
	ch       *amqp.Channel
	delayedQ *amqp.Queue
	taskQ    *amqp.Queue
	uriQ     *amqp.Queue
}

func NewRMQReceiverSender(cfg *DwhQueueServiceConfig) (*RMQReceiverSender, error) {
	addr := QueueAddrStringFromConfig(cfg)

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	maxPriorityArgs := map[string]interface{}{"x-max-priority": cfg.UriQueueMaxPriority}
	// Uri queue
	err = ch.ExchangeDeclare(
		cfg.ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		maxPriorityArgs,
	)

	maxPriorityAndTTLArgs := map[string]interface{}{
		"x-max-priority": cfg.UriQueueMaxPriority,
		"x-message-ttl":  cfg.DaemonTTLSeconds * 1000,
	}

	// Daemon task and delayed queues

	delayedQArgs := map[string]interface{}{
		"x-max-priority":            cfg.DaemonTaskQueueMaxPriority,
		"x-dead-letter-exchange":    cfg.ExchangeName,
		"x-dead-letter-routing-key": cfg.DaemonTaskQueueName,
		"x-message-ttl":             cfg.DaemonTTLSeconds * 1000,
	}

	uriQ, err := ch.QueueDeclare(
		cfg.UriQueueName,
		true,
		false,
		false,
		false,
		maxPriorityArgs,
	)
	if err != nil {
		return nil, err
	}

	delayedQ, err := ch.QueueDeclare(
		cfg.DaemonDelayedQueueName,
		true,
		false,
		false,
		false,
		delayedQArgs,
	)
	if err != nil {
		return nil, err
	}

	taskQ, err := ch.QueueDeclare(
		cfg.DaemonTaskQueueName,
		true,
		false,
		false,
		false,
		maxPriorityAndTTLArgs,
	)
	if err != nil {
		return nil, err
	}

	err = ch.Qos(
		cfg.DaemonTaskQueuePrefetchCount,
		0,
		false,
	)
	if err != nil {
		return nil, err
	}

	if err := ch.QueueBind(
		uriQ.Name,
		cfg.UriQueueName,
		cfg.ExchangeName,
		false,
		maxPriorityArgs,
	); err != nil {
		return nil, err
	}

	if err := ch.QueueBind(
		taskQ.Name,
		cfg.DaemonTaskQueueName,
		cfg.ExchangeName,
		false,
		maxPriorityArgs,
	); err != nil {
		return nil, err
	}

	if err := ch.QueueBind(
		delayedQ.Name,
		cfg.DaemonDelayedQueueName,
		cfg.ExchangeName,
		false,
		delayedQArgs,
	); err != nil {
		return nil, err
	}

	rs := &RMQReceiverSender{
		config:   cfg,
		conn:     conn,
		ch:       ch,
		uriQ:     &uriQ,
		delayedQ: &delayedQ,
		taskQ:    &taskQ,
	}

	// schedule future check
	err = rs.publishDelayed()
	if err != nil {
		return nil, err
	}

	return rs, nil

}

func (rs *RMQReceiverSender) Closer() error {
	if err := rs.ch.Close(); err != nil {
		return err
	}

	if err := rs.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (rs *RMQReceiverSender) GetTaskMessageChan() (<-chan amqp.Delivery, error) {
	msgs, err := rs.ch.Consume(
		rs.taskQ.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	return msgs, err
}

func (rs *RMQReceiverSender) PublishUriTask(url, owner, tokenId string) error {
	ba, err := json.Marshal(&TokenInfo{
		TokenID: tokenId,
		URL:     url,
		Owner:   owner,
	})
	if err != nil {
		return err
	}

	err = rs.ch.Publish(
		rs.config.ExchangeName,
		rs.uriQ.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         ba,
			Priority:     uint8(RegularUpdatePriority),
		})
	return err
}

func (rs *RMQReceiverSender) publishDelayed() error {
	err := rs.ch.Publish(
		rs.config.ExchangeName,
		rs.delayedQ.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte("ololo"),
		})
	return err
}
