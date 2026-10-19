package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Event struct {
	Type    string          `json:"type"`             // e.g. "notification", "task_updated"
	Payload json.RawMessage `json:"payload"`          // Raw JSON payload
	Target  string          `json:"target,omitempty"` // User ID or Project ID
}

func NewEvent(eventType string, payload any, target ...string) (*Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	tgt := ""
	if len(target) > 0 {
		tgt = target[0]
	}

	return &Event{
		Type:    eventType,
		Payload: data,
		Target:  tgt,
	}, nil
}

type Bus interface {
	Publish(ctx context.Context, channel string, event *Event) error
	SubscribePattern(ctx context.Context, pattern string) <-chan *redis.Message
}

type redisBus struct {
	client *redis.Client
}

func NewBus(client *redis.Client) Bus {
	return &redisBus{client: client}
}

func (b *redisBus) Publish(ctx context.Context, channel string, event *Event) error {
	if b.client == nil {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal pubsub event: %w", err)
	}

	return b.client.Publish(ctx, channel, data).Err()
}

func (b *redisBus) SubscribePattern(ctx context.Context, pattern string) <-chan *redis.Message {
	if b.client == nil {
		ch := make(chan *redis.Message)
		close(ch)
		return ch
	}

	pubsub := b.client.PSubscribe(ctx, pattern)
	return pubsub.Channel()
}
