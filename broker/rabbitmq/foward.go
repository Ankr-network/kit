package rabbitmq

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"strings"
)

type (
	RouteKeyConverter func(key string) string
	MsgConverter      func(msg amqp.Delivery) amqp.Publishing
)

func RetryError(srcUrl, dstUrl, errQ, dstEx string, nums int) {
	Forward(srcUrl, dstUrl, errQ, dstEx, nums, ErrRouteKeyConvert, SameMsgConvert)
}

func Copy(srcUrl, dstUrl, srcQ, dstEx string, nums int) {
	Forward(srcUrl, dstUrl, srcQ, dstEx, nums, SameRouteKeyConvert, SameMsgConvert)
}

func Forward(srcUrl, dstUrl string, srcQ, dstEx string, nums int, routeKeyConvert RouteKeyConverter, msgConvert MsgConverter) {
	source, err := amqp.Dial(srcUrl)
	if err != nil {
		log.Fatal("connection.open source error", zap.Error(err))
	}
	defer source.Close()

	chs, err := source.Channel()
	if err != nil {
		log.Fatal("channel.open source error", zap.Error(err))
	}

	shovel, err := chs.Consume(srcQ, "shovel", false, false, false, false, nil)
	if err != nil {
		log.Fatal("basic.consume source error", zap.Error(err))
	}

	// Setup the destination of the store and forward
	destination, err := amqp.Dial(dstUrl)
	if err != nil {
		log.Fatal("connection.open destination error", zap.Error(err))
	}
	defer destination.Close()

	chd, err := destination.Channel()
	if err != nil {
		log.Fatal("channel.open destination error", zap.Error(err))
	}

	// Buffer of 1 for our single outstanding publishing
	confirms := chd.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := chd.Confirm(false); err != nil {
		log.Fatal("confirm.select destination error", zap.Error(err))
	}

	// Now pump the messages, one by one, a smarter implementation
	// would batch the deliveries and use multiple ack/nacks
	counts := 0
	for {
		if counts == nums {
			log.Info("finished forward message", zap.Int("counts", counts))
			break
		}

		msg, ok := <-shovel
		if !ok {
			log.Info("source channel closed, see the reconnect example for handling this")
		}
		counts++

		err = chd.Publish(dstEx, routeKeyConvert(msg.RoutingKey), false, false, msgConvert(msg))

		if err != nil {
			msg.Nack(false, false)
			log.Fatal("basic.publish destination error", zap.Error(err))
		}

		log.Info("forward", zap.Int("counts", counts), zap.String("route_key", msg.RoutingKey), zap.Uint64("delivery_tag", msg.DeliveryTag))

		// only ack the source delivery when the destination acks the publishing
		if confirmed := <-confirms; confirmed.Ack {
			msg.Ack(false)
		} else {
			msg.Nack(false, false)
		}
	}
}

func ErrRouteKeyConvert(key string) string {
	return strings.TrimPrefix(key, "error.")
}

func SameRouteKeyConvert(key string) string {
	return key
}

func SameMsgConvert(msg amqp.Delivery) amqp.Publishing {
	return amqp.Publishing{
		// Copy all the properties
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,

		// Custom headers
		Headers: msg.Headers,

		// And the body
		Body: msg.Body,
	}
}
