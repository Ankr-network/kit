package rabbitmq

import (
	"errors"

	"github.com/Ankr-network/kit/broker"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

var (
	ErrPublishMessageNotAck = errors.New("message not ack by broker")
	ErrPublishMessageMiss   = errors.New("message cannot route to any queue")
)

type rabbitPublisher struct {
	broker   *rabbitBroker
	reliable bool
	topic    string
	conn     *Connection
}

func newRabbitPublisher(broker *rabbitBroker, topic string, reliable bool) (*rabbitPublisher, error) {
	out := &rabbitPublisher{
		broker:   broker,
		reliable: reliable,
		topic:    topic,
	}

	if err := out.init(); err != nil {
		return nil, err
	}

	return out, nil
}

func (rp *rabbitPublisher) init() error {
	conn, err := Dial(rp.broker.url)
	if err != nil {
		return err
	}

	rp.conn = conn

	return nil
}

func (rp *rabbitPublisher) Close() error {
	return rp.conn.Close()
}

func (rp *rabbitPublisher) Publish(m interface{}) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return ErrMessageIsNotProtoMessage
	}
	return rp.doPublish(rp.topic, msg)
}

func (rp *rabbitPublisher) PublishMessage(msg *broker.Message) error {
	return rp.doPublish(msg.Topic, msg.Value)
}

func (rp *rabbitPublisher) doPublish(topic string, msg proto.Message) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	publishing := amqp.Publishing{
		ContentType: "application/protobuf",
		Body:        body,
	}

	// ever single channel for publish
	ch, err := rp.conn.Channel(false)
	if err != nil {
		return err
	}
	defer ch.Close()

	if rp.reliable {
		if err := ch.Confirm(false); err != nil {
			return err
		}

		confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
		returnCh := ch.NotifyReturn(make(chan amqp.Return, 1))

		go func() {
			r, ok := <-returnCh
			if ok {
				logger.Errorf("message return: %+v", r)
			}
		}()

		publishing.DeliveryMode = amqp.Persistent

		if err := ch.Publish(rp.broker.exchange, topic, true, false, publishing); err != nil {
			return err
		}

		select {
		case <-returnCh:
			return ErrPublishMessageMiss
		case c := <-confirmCh:
			if !c.Ack {
				return ErrPublishMessageNotAck
			}
		}
	} else {
		if err := ch.Publish(rp.broker.exchange, topic, false, false, publishing); err != nil {
			return err
		}
	}

	return nil
}
