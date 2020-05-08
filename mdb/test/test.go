package test

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func DoTestInTx(cli *mongo.Client, f func(ctx mongo.SessionContext)) {
	cli.UseSession(context.Background(), func(ctx mongo.SessionContext) error {
		ctx.StartTransaction()
		defer ctx.AbortTransaction(ctx)
		f(ctx)
		return nil
	})
}

func Cleanup(col *mongo.Collection) {
	col.DeleteMany(context.Background(), bson.M{})
}
