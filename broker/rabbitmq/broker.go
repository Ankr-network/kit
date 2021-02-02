// Package rabbitmq provides a RabbitMQ broker
package rabbitmq

import (
	"fmt"
	"github.com/Ankr-network/kit/broker"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"regexp"
	"time"
)

var (
	rabbitUrlRegex = regexp.MustCompile(`^amqp(s)?://.+`)
)

type Options struct {
	DLX       string
	ALT       string
	NackDelay time.Duration
}

type Option func(opts *Options)

func WithDLX(dlx string) Option {
	return func(cfg *Options) {
		cfg.DLX = dlx
	}
}

func WithALT(alt string) Option {
	return func(cfg *Options) {
		cfg.ALT = alt
	}
}

func WithNackDelay(delay time.Duration) Option {
	return func(cfg *Options) {
		cfg.NackDelay = delay
	}
}

type rabbitBroker struct {
	url       string
	exchange  string
	dlx       string
	alt       string
	nackDelay time.Duration
}

func NewRabbitMQBrokerWithConfig() broker.Broker {
	return NewRabbitMQBroker(
		MustLoadConfig().URL,
		MustLoadConfig().Exchange,
		WithALT(MustLoadConfig().ALT),
		WithDLX(MustLoadConfig().DLX),
		WithNackDelay(MustLoadConfig().NackDelay),
	)
}

func NewRabbitMQBroker(url, exchange string, opts ...Option) broker.Broker {
	options := &Options{
		NackDelay: 5 * time.Second,
	}
	for _, o := range opts {
		o(options)
	}

	if !rabbitUrlRegex.MatchString(url) {
		log.Fatal("invalid RabbitMQ url", zap.String("url", url))
	}

	out := &rabbitBroker{
		url:       url,
		exchange:  exchange,
		dlx:       options.DLX,
		alt:       options.ALT,
		nackDelay: options.NackDelay,
	}

	out.init()

	return out
}

func (r *rabbitBroker) init() {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		log.Fatal("amqp.Dial error", zap.Error(err))
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("conn.Channel error", zap.Error(err))
	}
	defer ch.Close()

	if r.dlx != "" {
		mustTopicExchangeDeclare(r.dlx, nil, ch)
	}
	var exchangeArgs amqp.Table
	if r.alt != "" {
		mustTopicExchangeDeclare(r.alt, nil, ch)
		exchangeArgs = amqp.Table{"alternate-exchange": r.alt}
	}
	mustTopicExchangeDeclare(r.exchange, exchangeArgs, ch)
}

func (r *rabbitBroker) TopicPublisher(topic string, opts ...broker.Option) (broker.Publisher, error) {
	return r.createPublisher(topic, opts...)
}

func (r *rabbitBroker) MultiTopicPublisher(opts ...broker.Option) (broker.MultiTopicPublisher, error) {
	return r.createPublisher("", opts...)
}

func (r *rabbitBroker) RegisterSubscribeHandler(name, topic string, handler interface{}, opts ...broker.Option) error {
	brokerOptions := &broker.Options{
		Reliable: false,
		MaxRetry: 0,
	}

	for _, o := range opts {
		o(brokerOptions)
	}

	h, err := newHandler(handler, brokerOptions.Reliable, brokerOptions.MaxRetry, r.nackDelay)
	if err != nil {
		return err
	}

	s, err := newRabbitSubscriber(r, name, topic, brokerOptions.Reliable)
	if err != nil {
		return err
	}

	deliveries, err := s.Consume()
	if err != nil {
		return err
	}

	go h.consume(deliveries)

	return nil
}

func (r *rabbitBroker) RegisterErrSubscribeHandler(name, topic string, handler interface{}) error {
	if r.dlx == "" {
		return fmt.Errorf("broker without dead-letter exchange")
	}
	h, err := newErrHandler(handler)
	if err != nil {
		return err
	}

	s, err := newErrRabbitSubscriber(r, name, topic)
	if err != nil {
		return err
	}

	deliveries, err := s.Consume()
	if err != nil {
		return err
	}

	go h.consume(deliveries)

	return nil
}

func (r *rabbitBroker) createPublisher(topic string, opts ...broker.Option) (*rabbitPublisher, error) {
	brokerOptions := &broker.Options{
		Reliable: false,
		MaxRetry: 0,
	}

	for _, o := range opts {
		o(brokerOptions)
	}

	return newRabbitPublisher(r, topic, brokerOptions.Reliable)
}

// *** below are deprecated ***

func (r *rabbitBroker) Publisher(topic string, reliable bool) (broker.Publisher, error) {
	if reliable {
		return r.TopicPublisher(topic, broker.Reliable())
	} else {
		return r.TopicPublisher(topic)
	}
}

func (r *rabbitBroker) Subscribe(name, topic string, reliable, requeue bool, handler interface{}) error {
	if reliable {
		maxRetry := 0
		if requeue {
			maxRetry = 1
		}
		return r.RegisterSubscribeHandler(name, topic, handler, broker.Reliable(), broker.MaxRetry(maxRetry))
	} else {
		return r.RegisterSubscribeHandler(name, topic, handler)
	}
}
