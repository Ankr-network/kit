package mdb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

var (
	testCli *mongo.Client
)

func TestMain(m *testing.M) {
	testCli = NewClient(MustLoadConfig().URL)
	defer testCli.Disconnect(context.Background())
	m.Run()
}
