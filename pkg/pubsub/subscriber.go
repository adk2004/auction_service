package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

type topicState struct {
	pubsub  *redis.PubSub
	clients []chan Event
}

type Subscriber struct {
	mu     sync.RWMutex
	topics map[string]*topicState
	rdb    *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func NewSubscriber() *Subscriber {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &Subscriber{
		rdb:    rdb,
		topics: make(map[string]*topicState),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Subscriber) Subscribe(topic string) chan Event {
	c.mu.Lock()
	defer c.mu.Unlock()

	cl := make(chan Event, 32)
	// if the topic is already subscribed to by subscriber
	if state, exists := c.topics[topic]; exists {
		state.clients = append(state.clients, cl)
		return cl
	}
	// if the topic is not yet subscribed to by subscriber
	pubsub := c.rdb.Subscribe(c.ctx, topic)
	state := &topicState{
		pubsub:  pubsub,
		clients: []chan Event{cl},
	}
	c.topics[topic] = state

	go c.fanOut(topic, pubsub.Channel())

	return cl
}


func (c *Subscriber) fanOut(topic string, sch <-chan *redis.Message) {
	for {
		select {
		case msg, ok := <-sch:
			if !ok {
				c.CloseAllClients(topic)
				return
			}
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("pubsub: failed to unmarshal message on %q: %v", topic, err)
				continue
			}
			c.broadcast(topic, event)
		
		case <-c.ctx.Done():
			c.unsubscribeTopic(topic)
			c.CloseAllClients(topic)
			return
		}
	}
}

// broadcast sends the event to every registered client for the topic.
func (c *Subscriber) broadcast(topic string, event Event) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, ok := c.topics[topic]
	if !ok {
		return
	}

	for _, cl := range state.clients {
		select {
		case cl <- event:
		default:
			// Consumer is too slow; drop the message for this subscriber.
			// this does not slows other users as sending is non-blocking
			log.Printf("pubsub: slow consumer on %q, dropping message", topic)
		}
	}
}

func (c *Subscriber) Unsubscribe(topic string, cl chan Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.topics[topic]
	if !ok {
		return
	}
	for i, ch := range state.clients {
		if ch == cl {
			state.clients = append(state.clients[:i], state.clients[i+1:]...)
			close(cl)
			break
		}
	}

	if len(state.clients) == 0 {
		c.unsubscribeTopic(topic)
		delete(c.topics, topic)
	}
}


func (c *Subscriber) unsubscribeTopic(topic string) {
	state, ok := c.topics[topic]
	if !ok {
		return
	}
	if err := state.pubsub.Unsubscribe(c.ctx, topic); err != nil {
		log.Printf("pubsub: error unsubscribing from %q: %v", topic, err)
	}
	state.pubsub.Close()
}


func (c *Subscriber) CloseAllClients(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.topics[topic]
	if !ok {
		return
	}

	for _, cl := range state.clients {
		close(cl)
	}
	delete(c.topics, topic)
}

func (c *Subscriber) Close() error {
	c.cancel()
	return c.rdb.Close()
}