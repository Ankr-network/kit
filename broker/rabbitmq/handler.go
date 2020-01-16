package rabbitmq

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"log"
	"reflect"
	"time"
)

var (
	protoMessageType            = reflect.TypeOf((*proto.Message)(nil)).Elem()
	errorType                   = reflect.TypeOf((*error)(nil)).Elem()
	errTypeIsNotPtr             = errors.New("type must be pointer")
	errTypeIsNotProtoMessage    = errors.New("type must be proto.Message")
	errTypeIsNotError           = errors.New("type must be error")
	ErrMessageIsNotProtoMessage = errors.New("message must be proto.Message")
	ErrInvalidHandler           = errors.New("invalid handler, must be func and have single proto.Message implement and return single error")
	ErrMaxRetryTooMuch          = errors.New("currently only support maxRetry is 1")
)

type handler struct {
	methodValue reflect.Value
	msgType     reflect.Type
	reliable    bool
	maxRetry    int
	nackDelay   time.Duration
}

func newErrHandler(h interface{}) (*handler, error) {
	return newHandler(h, false, 0, 0)
}

func newHandler(h interface{}, reliable bool, maxRetry int, nacDelay time.Duration) (*handler, error) {
	if maxRetry > 1 {
		return nil, ErrMaxRetryTooMuch
	}
	ht := reflect.TypeOf(h)
	if ht.Kind() != reflect.Func {
		return nil, ErrInvalidHandler
	}

	if ht.NumIn() != 1 {
		return nil, ErrInvalidHandler
	}

	if ht.NumOut() != 1 {
		return nil, ErrInvalidHandler
	}

	mt := ht.In(0)
	if err := checkIsProtoMessage(mt); err != nil {
		return nil, ErrInvalidHandler
	}

	et := ht.Out(0)
	if err := checkIsError(et); err != nil {
		return nil, ErrInvalidHandler
	}

	return &handler{
		methodValue: reflect.ValueOf(h),
		msgType:     mt,
		reliable:    reliable,
		maxRetry:    maxRetry,
		nackDelay:   nacDelay,
	}, nil
}

func (h *handler) newMessage() proto.Message {
	return reflect.New(h.msgType.Elem()).Interface().(proto.Message)
}

func (h *handler) call(msg proto.Message) error {
	in := []reflect.Value{reflect.ValueOf(msg)}
	out := h.methodValue.Call(in)
	if out[0].IsNil() {
		return nil
	}
	return out[0].Interface().(error)
}

func (h *handler) consume(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		msg := h.newMessage()
		if err := proto.Unmarshal(d.Body, msg); err != nil {
			logger.Printf("proto.Unmarshal error: %v, %s", err, d.Body)
			if err := d.Nack(false, false); err != nil {
				log.Printf("Nack error: %v", err)
			}
			continue
		}

		if err := h.call(msg); err != nil {
			logger.Printf("handle message error: %v, message: %v", err, msg)
			if h.reliable {
				time.Sleep(h.nackDelay)

				if d.Redelivered {
					if err := d.Nack(false, false); err != nil {
						log.Printf("Nack error: %v", err)
					}
				} else {
					if err := d.Nack(false, h.maxRetry > 0); err != nil {
						log.Printf("Nack error: %v", err)
					}
				}
			}
		} else {
			if h.reliable {
				if err := d.Ack(false); err != nil {
					logger.Printf("Ack error: %v", err)
				}
			}
		}
	}
}

func checkIsProtoMessage(t reflect.Type) error {
	if t.Kind() != reflect.Ptr {
		return errTypeIsNotPtr
	}

	if !t.Implements(protoMessageType) {
		return errTypeIsNotProtoMessage
	}
	return nil
}

func checkIsError(t reflect.Type) error {
	if !t.Implements(errorType) {
		return errTypeIsNotError
	}
	return nil
}
