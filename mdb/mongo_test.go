//+build integration

package mdb

import (
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"kit/mdb/test"
	"testing"
)

// set MONGO_URL env
func TestAggregateInTx(t *testing.T) {
	col := testCli.Database("test").Collection("tx")

	type foo struct {
		Bar string
	}

	test.DoTestInTx(testCli, func(ctx mongo.SessionContext) {
		_, err := col.InsertOne(ctx, &foo{Bar: "bar"})
		require.NoError(t, err)

		var out1 foo
		err = col.FindOne(ctx, bson.M{}).Decode(&out1)
		require.NoError(t, err)
		t.Logf("FindOne foo: %+v", out1)

		cur, err := col.Aggregate(ctx, []bson.M{
			{"$match": bson.M{"bar": bson.M{"$exists": true}}},
		})
		require.NoError(t, err)
		var out2 []*foo
		err = cur.All(context.TODO(), &out2)
		require.NoError(t, err)
		for _, f := range out2 {
			t.Logf("Aggregate foo: %+v", f)
		}
	})
}

// set MONGO_URL env
func TestDecimalCodec(t *testing.T) {
	col := testCli.Database("test").Collection("decimal")

	type foo struct {
		ID  string          `bson:"id"`
		Bar decimal.Decimal `bson:"bar"`
	}

	test.DoTestInTx(testCli, func(ctx mongo.SessionContext) {
		in := foo{ID: uuid.New().String(), Bar: decimal.New(1024, -2)}
		_, err := col.InsertOne(ctx, &in)
		require.NoError(t, err)

		var db foo
		err = col.FindOne(ctx, bson.M{"id": in.ID}).Decode(&db)
		assert.Equal(t, in, db)
		assert.NoError(t, err)
	})
}

// set MONGO_URL env
func TestIsDuplicateKeyError(t *testing.T) {
	col := testCli.Database("test").Collection("dup")

	type foo struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	test.DoTestInTx(testCli, func(ctx mongo.SessionContext) {
		in := foo{}
		_, err := col.InsertOne(ctx, &in)
		require.NoError(t, err)
		_, err = col.InsertOne(ctx, &in)
		assert.True(t, IsDuplicateKeyError(err))
	})
}

func TestIsDuplicateKeyErrorWithoutTx(t *testing.T) {
	col := testCli.Database("test").Collection("dup")
	defer test.Cleanup(col)

	type foo struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	ctx := context.Background()

	in := foo{}
	_, err := col.InsertOne(ctx, &in)
	require.NoError(t, err)
	_, err = col.InsertOne(ctx, &in)
	assert.True(t, IsDuplicateKeyError(err))
}
