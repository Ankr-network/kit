package rabbitmq

import (
	"github.com/streadway/amqp"
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
		logger.Fatalf("connection.open source: %s", err)
	}
	defer source.Close()

	chs, err := source.Channel()
	if err != nil {
		logger.Fatalf("channel.open source: %s", err)
	}

	shovel, err := chs.Consume(srcQ, "shovel", false, false, false, false, nil)
	if err != nil {
		logger.Fatalf("basic.consume source: %s", err)
	}

	// Setup the destination of the store and forward
	destination, err := amqp.Dial(dstUrl)
	if err != nil {
		logger.Fatalf("connection.open destination: %s", err)
	}
	defer destination.Close()

	chd, err := destination.Channel()
	if err != nil {
		logger.Fatalf("channel.open destination: %s", err)
	}

	// Buffer of 1 for our single outstanding publishing
	confirms := chd.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := chd.Confirm(false); err != nil {
		logger.Fatalf("confirm.select destination: %s", err)
	}

	// Now pump the messages, one by one, a smarter implementation
	// would batch the deliveries and use multiple ack/nacks
	counts := 0
	for {
		if counts == nums {
			logger.Infof("finished forward %d message", nums)
			break
		}

		msg, ok := <-shovel
		if !ok {
			logger.Fatalf("source channel closed, see the reconnect example for handling this")
		}
		counts++

		err = chd.Publish(dstEx, routeKeyConvert(msg.RoutingKey), false, false, msgConvert(msg))

		if err != nil {
			msg.Nack(false, false)
			logger.Fatalf("basic.publish destination: %+v", msg)
		}

		logger.Infof("forward %d %s:%d", counts, msg.RoutingKey, msg.DeliveryTag)

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
