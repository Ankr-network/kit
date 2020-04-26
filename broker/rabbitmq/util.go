package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func mustTopicExchangeDeclare(name string, args amqp.Table, channel *amqp.Channel) {
	if err := channel.ExchangeDeclare(name, "topic", true, false, false, false, args); err != nil {
		log.Fatal("channel.ExchangeDeclare error: %v", zap.String("exchange", name), zap.Error(err))
	}
}

func queueBind(queue, key, exchange string, channel *amqp.Channel) error {
	if err := channel.QueueBind(queue, key, exchange, false, nil); err != nil {
		return err
	}
	return nil
}

func queueDeclare(name, topic, dlx string, reliable bool, conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	args := amqp.Table{}
	if !reliable { // not reliable
		args["x-message-ttl"] = 20000 // 20 second
	} else if dlx != "" { // reliable with dlx
		args["x-dead-letter-exchange"] = dlx
		args["x-dead-letter-routing-key"] = fmt.Sprintf("error.%s", topic)
	}
	if _, declareErr := ch.QueueDeclare(name, true, false, false, false, args); declareErr != nil {
		log.Info("try recreate queue for channel.QueueDeclare error", zap.Error(declareErr))
		ach, err := conn.Channel()
		if err != nil {
			return err
		}
		defer ach.Close()
		_, err = ach.QueueDelete(name, false, true, false)
		if err != nil {
			log.Error("channel.QueueDelete error", zap.Error(err))
			return declareErr
		}
		_, err = ach.QueueDeclare(name, true, false, false, false, args)
		if err != nil {
			log.Error("channel.QueueDeclare again error", zap.Error(err))
			return declareErr
		}
	}
	return nil
}
