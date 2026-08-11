package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Publisher struct {
	Brokers                 []string
	Log                     *slog.Logger
	AllowNoopWithoutBrokers bool
	mu                      sync.Mutex
	writers                 map[string]*kafkago.Writer
}

func NewPublisher(brokers []string, log *slog.Logger) *Publisher {
	if log == nil {
		log = slog.Default()
	}
	return &Publisher{Brokers: brokers, Log: log, writers: make(map[string]*kafkago.Writer)}
}

func (p *Publisher) writer(topic string) *kafkago.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(p.Brokers...),
		Topic:        topic,
		Balancer:     &kafkago.Hash{},
		RequiredAcks: kafkago.RequireOne,
		BatchTimeout: 10 * time.Millisecond,
	}
	p.writers[topic] = w
	return w
}

func (p *Publisher) Publish(ctx context.Context, topic, key string, payload map[string]any) error {
	if len(p.Brokers) == 0 {
		if p.AllowNoopWithoutBrokers {
			p.Log.Debug("kafka.noop", "topic", topic, "key", key)
			return nil
		}
		return fmt.Errorf("kafka: no brokers configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := p.writer(topic).WriteMessages(ctx, kafkago.Message{Key: []byte(key), Value: body, Time: time.Now().UTC()}); err != nil {
		return fmt.Errorf("kafka: publish %s: %w", topic, err)
	}
	p.Log.Info("kafka.publish", "topic", topic, "key", key, "bytes", len(body))
	return nil
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var first error
	for t, w := range p.writers {
		if err := w.Close(); err != nil && first == nil {
			first = err
		}
		delete(p.writers, t)
	}
	return first
}
