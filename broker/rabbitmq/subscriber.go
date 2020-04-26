package rabbitmq

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type rabbitSubscriber struct {
	broker   *rabbitBroker
	name     string
	topic    string
	reliable bool
	conn     *Connection
	channel  *Channel
	isErrSub bool
}

func newRabbitSubscriber(broker *rabbitBroker, name, topic string, reliable bool) (*rabbitSubscriber, error) {
	out := &rabbitSubscriber{
		broker:   broker,
		name:     name,
		topic:    topic,
		reliable: reliable,
		isErrSub: false,
	}
	if err := out.init(); err != nil {
		return nil, err
	}

	return out, nil
}

func newErrRabbitSubscriber(broker *rabbitBroker, name, errTopic string) (*rabbitSubscriber, error) {
	out := &rabbitSubscriber{
		broker:   broker,
		name:     name,
		topic:    errTopic,
		reliable: true,
		isErrSub: true,
	}
	if err := out.init(); err != nil {
		return nil, err
	}

	return out, nil
}

func (rs *rabbitSubscriber) init() error {
	conn, err := Dial(rs.broker.url)
	if err != nil {
		return err
	}
	conn.m.RLock()
	ch, err := conn.Channel(true)
	conn.m.RUnlock()
	if err != nil {
		if err := conn.Close(); err != nil {
			log.Error("conn.Close error", zap.Error(err))
		}
		return err
	}

	var dlx string
	if rs.reliable && !rs.isErrSub {
		dlx = rs.broker.dlx
	}

	if err := queueDeclare(rs.name, rs.topic, dlx, rs.reliable, conn.Connection); err != nil {
		if err := conn.Close(); err != nil {
			log.Error("conn.Close error", zap.Error(err))
		}
		return err
	}

	var exchange string
	if rs.isErrSub {
		exchange = rs.broker.dlx
	} else {
		exchange = rs.broker.exchange
	}

	ch.m.RLock()
	if err := queueBind(rs.name, rs.topic, exchange, ch.Channel); err != nil {
		ch.m.RUnlock()
		if err := conn.Close(); err != nil {
			log.Error("conn.Close error", zap.Error(err))
		}
		return err
	}
	ch.m.RUnlock()

	rs.conn = conn
	rs.channel = ch

	return nil
}

func (rs *rabbitSubscriber) Close() error {
	return rs.conn.Close()
}

func (rs *rabbitSubscriber) Consume() (<-chan amqp.Delivery, error) {
	autoAck := true
	if rs.reliable && !rs.isErrSub {
		autoAck = false
	}
	rs.channel.m.RLock()
	deliveries, err := rs.channel.Consume(rs.name, "", autoAck, false, false, false, nil)
	rs.channel.m.RUnlock()
	return deliveries, err

}
