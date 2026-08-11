package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/segmentio/kafka-go"
)

// Publisher is a production Kafka event publisher (segmentio/kafka-go).
// When Brokers is empty, Publish returns an error so production never silently drops events.
// Dev/test callers may set AllowNoopWithoutBrokers=true.
type Publisher struct {
	Brokers                 []string
	Log                     *slog.Logger
	AllowNoopWithoutBrokers bool
	mu                      sync.Mutex
	writers                 map[string]*kafka.Writer
}

func NewPublisher(brokers []string, log *slog.Logger) *Publisher {
	if log == nil {
		log = slog.Default()
	}
	return &Publisher{
		Brokers: append([]string(nil), brokers...),
		Log:     log,
		writers: make(map[string]*kafka.Writer),
	}
}

func (p *Publisher) writer(topic string) *kafka.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		BatchTimeout: 10 * time.Millisecond,
	}
	p.writers[topic] = w
	return w
}

func (p *Publisher) Publish(ctx context.Context, topic, key string, payload any) error {
	if len(p.Brokers) == 0 {
		if p.AllowNoopWithoutBrokers {
			p.Log.Debug("kafka.noop", "topic", topic, "key", key)
			return nil
		}
		return fmt.Errorf("kafka: no brokers configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("kafka: marshal: %w", err)
	}
	w := p.writer(topic)
	err = w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("kafka: publish %s: %w", topic, err)
	}
	p.Log.Info("kafka.publish", "topic", topic, "key", key, "bytes", len(body))
	return nil
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var first error
	for topic, w := range p.writers {
		if err := w.Close(); err != nil && first == nil {
			first = err
		}
		delete(p.writers, topic)
	}
	return first
}

var _ ports.EventPublisher = (*Publisher)(nil)
