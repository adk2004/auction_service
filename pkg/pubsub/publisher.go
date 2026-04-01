package pubsub

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	rdb *redis.Client
}

func NewPublisher() *Publisher {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", 
		DB:       0,
	})
	return &Publisher{
		rdb: rdb,
	}
}

func (p *Publisher) Publish(ctx context.Context, topic string, event Event) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.rdb.Publish(ctx, topic, string(msg)).Err()
}

func (p *Publisher) Close() error {
	return p.rdb.Close()
}