package main

import (
	"fmt"
	"github.com/Ankr-network/kit/broker/example/proto"
	"github.com/Ankr-network/kit/broker/rabbitmq"
	"log"
	"time"

	"github.com/Ankr-network/kit/broker"
)

var (
	topic            = "ankr.topic.hello"
	ankrBroker       broker.Broker
	helloPublisher   broker.Publisher
	helloSubscriber1 = logHandler{name: "hello1"}
)

type logHandler struct {
	name string
}

func (s *logHandler) handle(h *proto.Hello) error {
	log.Printf("%s handle %+v", s.name, h)
	return nil
}

func init() {
	var err error
	ankrBroker = rabbitmq.NewBroker()
	if helloPublisher, err = ankrBroker.Publisher(topic, true); err != nil {
		log.Fatal(err)
	}
	if err := ankrBroker.Subscribe("hello1", topic, true, false, helloSubscriber1.handle); err != nil {
		log.Fatal(err)
	}
}

func pub() {
	tick := time.NewTicker(time.Second)
	i := 0
	for range tick.C {
		msg := proto.Hello{Name: fmt.Sprintf("No.%d", i)}
		if err := helloPublisher.Publish(&msg); err != nil {
			log.Printf("[pub] failed: %v", err)
		} else {
			log.Printf("[pub] pubbed message: %v", msg)
		}
		i++
	}
}

func main() {
	go pub()
	<-time.After(time.Second * 100)
}
