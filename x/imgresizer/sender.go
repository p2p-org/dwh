package imgresizer

import (
	"encoding/json"

	dwh_common "github.com/dgamingfoundation/dwh/x/common"

	"github.com/streadway/amqp"
)

type RMQSender struct {
	config *dwh_common.DwhCommonServiceConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	imgQ   *amqp.Queue
}

func NewRMQSender(configFileName, configPath string) (*RMQSender, error) {
	rCfg := dwh_common.ReadCommonConfig(configFileName, configPath)
	u := dwh_common.QueueAddrStringFromConfig(rCfg)

	conn, err := amqp.Dial(u)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	qArgs := map[string]interface{}{"x-max-priority": rCfg.ImgQueueMaxPriority}
	q, err := ch.QueueDeclare(
		rCfg.ImgQueueName,
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
		config: rCfg,
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

func (rs *RMQSender) Publish(imgUrl, owner, tokenId string, priority dwh_common.ImgQueuePriority) error {
	ba, err := json.Marshal(&dwh_common.TaskInfo{
		Owner:   owner,
		URL:     imgUrl,
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
