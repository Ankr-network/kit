package mdb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(client *mongo.Client, dbName, collectionName string) *Repository {
	return &Repository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

func NewClientWithConfig() *mongo.Client {
	return NewClient(MustLoadConfig().URL)
}

func NewClient(url string) *mongo.Client {
	rb := bson.NewRegistryBuilder()
	codec := &DecimalCodec{}
	rb.RegisterCodec(tDecimal, codec)
	register := rb.Build()
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url).SetRegistry(register))
	if err != nil {
		log.Fatal("mongo.Connect error", zap.Error(err))
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("client.Ping error", zap.Error(err))
	}
	return client
}

func (m *Repository) AddOne(ctx context.Context, entity interface{}) error {
	if _, err := m.collection.InsertOne(ctx, entity); err != nil {
		if IsDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		log.Error("InsertOne error", zap.Error(err))
		return err
	}

	return nil
}

func (m *Repository) SaveOne(ctx context.Context, filter bson.M, entity interface{}) error {
	if _, err := m.collection.ReplaceOne(ctx, filter, entity, options.Replace().SetUpsert(true)); err != nil {
		if IsDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		log.Error("ReplaceOne error", zap.Error(err))
		return err
	}

	return nil
}

func (m *Repository) FindOne(ctx context.Context, filter bson.M, result interface{}, opts ...*options.FindOneOptions) error {
	if err := m.collection.FindOne(ctx, filter, opts...).Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		log.Error("FindOne error", zap.Error(err))
		return err
	}
	return nil
}

func (m *Repository) Find(ctx context.Context, filter bson.M, results interface{}, opts ...*options.FindOptions) error {
	cur, err := m.collection.Find(ctx, filter, opts...)
	if err != nil {
		log.Error("Find error", zap.Error(err))
		return err
	}

	if err := cur.All(ctx, results); err != nil {
		log.Error("Cursor.All error", zap.Error(err))
		return err
	}
	return nil
}

func (m *Repository) DeleteOne(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) error {
	delRes, err := m.collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		log.Error("DeleteOne error", zap.Error(err))
		return err
	}

	if delRes.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *Repository) DeleteMany(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) error {
	_, err := m.collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		log.Error("DeleteMany error", zap.Error(err))
		return err
	}

	return nil
}

func (m *Repository) UpdateOne(ctx context.Context, filter bson.M, entity interface{}) error {
	updateRes, err := m.collection.ReplaceOne(ctx, filter, entity)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		log.Error("ReplaceOne error", zap.Error(err))
		return err
	}

	if updateRes.ModifiedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *Repository) GetCollection() *mongo.Collection {
	return m.collection
}

func (m *Repository) GetClient() *mongo.Client {
	return m.collection.Database().Client()
}
