// Package rabbitmq provides a RabbitMQ broker
package rabbitmq

import (
	"github.com/Ankr-network/kit/broker"
	"github.com/streadway/amqp"
	"regexp"
	"time"
)

var (
	rabbitUrlRegex = regexp.MustCompile(`^amqp(s)?://.+`)
)

type Option func(cfg *Config)

func WithAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.URL = addr
	}
}

func WithExchange(exchange string) Option {
	return func(cfg *Config) {
		cfg.Exchange = exchange
	}
}

func WithDLX(dlx string) Option {
	return func(cfg *Config) {
		cfg.DLX = dlx
	}
}

type rabbitBroker struct {
	url       string
	exchange  string
	dlx       string
	alt       string
	nackDelay time.Duration
}

func NewRabbitMQBroker(opts ...Option) broker.Broker {
	cfg, _ := LoadConfig()

	for _, o := range opts {
		o(cfg)
	}

	if !rabbitUrlRegex.MatchString(cfg.URL) {
		logger.Fatalf("invalid RabbitMQ url: %s", cfg.URL)
	}

	out := &rabbitBroker{
		url:       cfg.URL,
		exchange:  cfg.Exchange,
		dlx:       cfg.DLX,
		alt:       cfg.ALT,
		nackDelay: cfg.NackDelay,
	}

	out.init()

	return out
}

func (r *rabbitBroker) init() {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		logger.Fatalf("amqp.Dial error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("conn.Channel error: %v", err)
	}
	defer ch.Close()

	if err := topicExchangeDeclare(r.alt, nil, ch); err != nil {
		logger.Fatalf("topicExchangeDeclare %s error: %v", r.alt, err)
	}
	if err := topicExchangeDeclare(r.dlx, nil, ch); err != nil {
		logger.Fatalf("topicExchangeDeclare %s error: %v", r.dlx, err)
	}
	if err := topicExchangeDeclare(r.exchange, amqp.Table{"alternate-exchange": r.alt}, ch); err != nil {
		logger.Fatalf("topicExchangeDeclare %s error: %v", r.exchange, err)
	}
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

// Deprecated. use NewRabbitMQBroker instead
func NewBroker(args ...string) broker.Broker {
	if len(args) > 0 {
		return NewRabbitMQBroker(WithAddr(args[0]))
	} else {
		return NewRabbitMQBroker()
	}
}
