package main

import (
	"errors"
	"fmt"
	"github.com/Ankr-network/kit/broker/example/proto"
	"github.com/Ankr-network/kit/broker/rabbitmq"
	"github.com/Ankr-network/kit/log"
	"time"

	"github.com/Ankr-network/kit/broker"
)

var (
	topic            = "ankr.topic.hello"
	dlqTopic         = "error.ankr.topic.hello"
	ankrBroker       broker.Broker
	publisher        broker.MultiTopicPublisher
	helloPublisher   broker.Publisher
	helloSubscriber1 = logHandler{name: "hello1"}
	logger           = log.Logger()
)

type logHandler struct {
	name string
}

func (s *logHandler) handle(h *proto.Hello) error {
	logger.Infof("%s handle %+v", s.name, h)
	return nil
}

func (s *logHandler) errHandle(h *proto.Hello) error {
	logger.Infof("%s errHandle %+v", s.name, h)
	return errors.New("some error")
}

func (s *logHandler) dlqHandle(h *proto.Hello) error {
	logger.Infof("%s dlqHandle %+v", s.name, h)
	return nil
}

func init() {
	var err error
	ankrBroker = rabbitmq.NewRabbitMQBrokerFromConfig()
	if publisher, err = ankrBroker.MultiTopicPublisher(broker.Reliable()); err != nil {
		logger.Fatal(err)
	}
	if helloPublisher, err = ankrBroker.TopicPublisher(topic, broker.Reliable()); err != nil {
		logger.Fatal(err)
	}
	if err := ankrBroker.RegisterSubscribeHandler("hello1", topic, helloSubscriber1.handle, broker.Reliable()); err != nil {
		logger.Fatal(err)
	}
	if err := ankrBroker.RegisterSubscribeHandler("hello1err", topic, helloSubscriber1.errHandle, broker.Reliable()); err != nil {
		logger.Fatal(err)
	}
	if err := ankrBroker.RegisterErrSubscribeHandler("hello1dlq", dlqTopic, helloSubscriber1.dlqHandle); err != nil {
		logger.Fatal(err)
	}
}

func multiPub() {
	tick := time.NewTicker(time.Second)
	i := 0
	for range tick.C {
		msg := proto.Hello{Name: fmt.Sprintf("No.%d", i)}
		if err := publisher.PublishMessage(&broker.Message{
			Topic: topic,
			Value: &msg,
		}); err != nil {
			logger.Infof("[multiPub] failed: %v", err)
		} else {
			logger.Infof("[multiPub] pubbed message: %+v", msg)
		}
		i++
	}
}

func pub() {
	tick := time.NewTicker(time.Second)
	i := 0
	for range tick.C {
		msg := proto.Hello{Name: fmt.Sprintf("No.%d", i)}
		if err := helloPublisher.Publish(&msg); err != nil {
			logger.Infof("[pub] failed: %v", err)
		} else {
			logger.Infof("[pub] pubbed message: %+v", msg)
		}
		i++
	}
}

func main() {
	go pub()
	<-time.After(time.Second * 100)
}
