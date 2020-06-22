// Package broker is an interface used for asynchronous messaging
package broker

import "github.com/golang/protobuf/proto"

// Broker is an interface used for asynchronous messaging.
type Broker interface {
	TopicPublisher(topic string, opts ...Option) (Publisher, error)
	MultiTopicPublisher(opts ...Option) (MultiTopicPublisher, error)
	RegisterSubscribeHandler(name, topic string, handler interface{}, opts ...Option) error
	RegisterErrSubscribeHandler(name, topic string, handler interface{}) error
}

type Publisher interface {
	Publish(m interface{}) error
}

type MultiTopicPublisher interface {
	PublishMessage(msg *Message) error
}

type Message struct {
	Topic string
	Value proto.Message
}

func NewMessage(topic string, value proto.Message) *Message {
	return &Message{
		Topic: topic,
		Value: value,
	}
}

type Options struct {
	Reliable bool
	MaxRetry int
}

type Option func(opts *Options)

func Reliable() Option {
	return func(opts *Options) {
		opts.Reliable = true
	}
}

func MaxRetry(retry int) Option {
	return func(opts *Options) {
		opts.MaxRetry = retry
	}
}
