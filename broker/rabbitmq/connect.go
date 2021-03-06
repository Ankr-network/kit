package rabbitmq

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

const (
	consumeRetryDelay        = 8
	channelReconnectDelay    = 4
	connectionReconnectDelay = 2
)

// Connection amqp.Connection wrapper
type Connection struct {
	*amqp.Connection
	m sync.RWMutex
}

// Dial wrap amqp.Dial, dial and get a reconnect connection
func Dial(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	connection := &Connection{
		Connection: conn,
	}

	go func() {
		for {
			reason, ok := <-connection.Connection.NotifyClose(make(chan *amqp.Error))
			// exit this goroutine if closed by developer
			if !ok {
				log.Info("connection closed")
				break
			}
			log.Info("connection closed", zap.Reflect("reason", reason))

			connection.m.Lock()
			// reconnect if not closed by developer
			for {
				// wait 1s for reconnect
				time.Sleep(connectionReconnectDelay * time.Second)

				conn, err := amqp.Dial(url)
				if err == nil {
					connection.Connection = conn
					log.Info("reconnect success")
					connection.m.Unlock()
					break
				}

				log.Error("reconnect error", zap.Error(err))
			}
		}
	}()

	return connection, nil
}

// Channel wrap amqp.Connection.Channel, get a auto reconnect channel
func (c *Connection) Channel(reconnect bool) (*Channel, error) {
	c.m.RLock()
	ch, err := c.Connection.Channel()
	c.m.RUnlock()
	if err != nil {
		return nil, err
	}

	resultChannel := &Channel{
		Channel: ch,
	}

	if reconnect {
		go func() {
			for {
				reason, ok := <-resultChannel.Channel.NotifyClose(make(chan *amqp.Error))
				// exit this goroutine if closed by developer
				if !ok || resultChannel.IsClosed() {
					log.Info("channel closed")
					if err := resultChannel.Close(); err != nil { // close again, ensure closed flag set when connection closed
						log.Error("Channel.Close error", zap.Error(err))
					}
					break
				}
				log.Info("channel closed", zap.Reflect("reason", reason))

				resultChannel.m.Lock()
				// reconnect if not closed by developer
				for {
					// wait 1s for connection reconnect
					time.Sleep(channelReconnectDelay * time.Second)

					c.m.RLock()
					ch, err := c.Connection.Channel()
					c.m.RUnlock()
					if err == nil {
						log.Info("channel recreate success")
						resultChannel.Channel = ch
						resultChannel.m.Unlock()
						break
					}
					log.Error("channel recreate error", zap.Error(err))
				}
			}
		}()
	}

	return resultChannel, nil
}

// Channel amqp.Channel wrapper
type Channel struct {
	*amqp.Channel
	closed int32
	m      sync.RWMutex
}

// IsClosed indicate closed by developer
func (ch *Channel) IsClosed() bool {
	return atomic.LoadInt32(&ch.closed) == 1
}

// Close ensure closed flag set
func (ch *Channel) Close() error {
	if ch.IsClosed() {
		return amqp.ErrClosed
	}

	atomic.StoreInt32(&ch.closed, 1)

	return ch.Channel.Close()
}

// Consume wrap amqp.Channel.Consume, the returned delivery will end only when channel closed by developer
func (ch *Channel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	deliveries := make(chan amqp.Delivery)

	go func() {
		for {
			ch.m.RLock()
			d, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
			ch.m.RUnlock()
			if err != nil {
				log.Error("consume error", zap.Error(err))
				time.Sleep(consumeRetryDelay * time.Second)
				continue
			}

			for msg := range d {
				deliveries <- msg
			}

			// sleep before IsClose call. closed flag may not set before sleep.
			time.Sleep(consumeRetryDelay * time.Second)

			if ch.IsClosed() {
				break
			}
		}
	}()

	return deliveries, nil
}
