// +build integration

package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPublishConfirmAndReturn(t *testing.T) {
	cfg, err := LoadConfig()
	require.NoError(t, err)
	conn, err := amqp.Dial(cfg.URL)
	require.NoError(t, err)
	defer conn.Close()
	ch, err := conn.Channel()
	defer ch.Close()

	publishCh := make(chan amqp.Confirmation, 1)
	ch.NotifyPublish(publishCh)
	returnCh := make(chan amqp.Return, 1)
	ch.NotifyReturn(returnCh)

	ch.Confirm(false)

	err = ch.Publish("test", "test.hello", true, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         []byte("hello"),
	})
	assert.NoError(t, err)

	select {
	case r := <-returnCh:
		t.Logf("return: %+v", r)
	case <-time.After(time.Second):
		t.Log("timeout")
	}

	c := <-publishCh
	t.Logf("confirm: %+v", c)
}
