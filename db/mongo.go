package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"kavya/ai"
)

type MongoStore struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type MongoMessageDoc struct {
	ChatID    int64     `bson:"chat_id"`
	Role      string    `bson:"role"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}

func NewMongoStore(uri, dbName string) (*MongoStore, error) {
	if dbName == "" {
		dbName = "kavya"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping error: %w", err)
	}

	coll := client.Database(dbName).Collection("messages")

	// Create index on chat_id and _id for fast queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "chat_id", Value: 1},
			{Key: "_id", Value: -1},
		},
	}
	_, _ = coll.Indexes().CreateOne(ctx, indexModel)

	return &MongoStore{
		client:     client,
		collection: coll,
	}, nil
}

func (m *MongoStore) SaveMessage(chatID int64, role, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := MongoMessageDoc{
		ChatID:    chatID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	_, err := m.collection.InsertOne(ctx, doc)
	return err
}

func (m *MongoStore) GetHistory(chatID int64, limit int) ([]ai.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 20
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := m.collection.Find(ctx, bson.M{"chat_id": chatID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []MongoMessageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	// Reverse to chronological order (oldest -> newest)
	history := make([]ai.Message, len(docs))
	for i, d := range docs {
		history[len(docs)-1-i] = ai.Message{
			Role:    d.Role,
			Content: d.Content,
		}
	}

	return history, nil
}

func (m *MongoStore) ClearHistory(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.DeleteMany(ctx, bson.M{"chat_id": chatID})
	return err
}

func (m *MongoStore) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}
