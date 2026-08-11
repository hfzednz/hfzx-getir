// Package kafka provides a production Kafka event publisher using segmentio/kafka-go.
//
// Behavior:
//   - When brokers are configured (KAFKA_BROKERS non-empty): Publish writes to Kafka.
//   - When brokers are empty: Publish is a documented no-op (returns nil) so local/dev
//     boots without a broker. Callers that require delivery must configure brokers.
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

// Publisher publishes domain events to Kafka when brokers are configured.
type Publisher struct {
	brokers []string
	log     *slog.Logger

	mu      sync.Mutex
	writers map[string]*kafkago.Writer
}

// NewPublisher returns a Kafka publisher.
// Empty brokers → no-op Publish (documented); non-empty → real segmentio/kafka-go writer.
func NewPublisher(brokers []string, log *slog.Logger) *Publisher {
	if log == nil {
		log = slog.Default()
	}
	return &Publisher{
		brokers: append([]string(nil), brokers...),
		log:     log,
		writers: make(map[string]*kafkago.Writer),
	}
}

// Publish emits payload as JSON to the given topic/key.
// No-op (nil error) when no brokers are configured.
func (p *Publisher) Publish(ctx context.Context, topic, key string, payload any) error {
	if len(p.brokers) == 0 {
		p.log.Debug("kafka.noop", "topic", topic, "key", key, "reason", "KAFKA_BROKERS empty")
		return nil
	}
	if topic == "" {
		return fmt.Errorf("kafka publish: empty topic")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("kafka marshal: %w", err)
	}
	w := p.writerFor(topic)
	msg := kafkago.Message{
		Key:   []byte(key),
		Value: body,
		Time:  time.Now().UTC(),
	}
	if err := w.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka write topic=%s: %w", topic, err)
	}
	p.log.Debug("kafka.published", "topic", topic, "key", key, "bytes", len(body))
	return nil
}

// Close flushes and closes all topic writers.
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var first error
	for topic, w := range p.writers {
		if err := w.Close(); err != nil && first == nil {
			first = fmt.Errorf("close writer %s: %w", topic, err)
		}
		delete(p.writers, topic)
	}
	return first
}

func (p *Publisher) writerFor(topic string) *kafkago.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
	}
	p.writers[topic] = w
	return w
}
