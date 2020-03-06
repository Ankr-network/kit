//+build integration

package rabbitmq

import "testing"

func TestRetryError(t *testing.T) {
	srcUrl := "amqp://guest:guest@127.0.0.1:5672"
	dstUrl := "amqp://guest:guest@127.0.0.1:5672"
	srcQ := "error"
	dstEx := "ankr.topic"

	RetryError(srcUrl, dstUrl, srcQ, dstEx, 1)
}

func TestCopy(t *testing.T) {
	srcUrl := "amqp://guest:guest@127.0.0.1:5672"
	dstUrl := "amqp://guest:guest@127.0.0.1:5672"
	srcQ := "miss"
	dstEx := "ankr.topic"

	Copy(srcUrl, dstUrl, srcQ, dstEx, 1)
}
