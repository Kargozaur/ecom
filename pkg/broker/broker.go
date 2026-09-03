package broker

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	Key   []byte
	Value []byte
}

type Writer struct {
	writer   *kafka.Writer
	messages []*Message
	mu       sync.RWMutex
	m        int
}

type Reader struct {
	reader *kafka.Reader
}

func NewKafkaReader(topic string, brokers []string, partition int) (*Reader, error) {
	if len(brokers) == 0 {
		return nil, ErrNoBrokers
	}
	if topic == "" {
		return nil, ErrTopicRequired
	}
	return &Reader{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:   brokers,
			Topic:     topic,
			Partition: partition,
			MaxBytes:  10e6, // 10 MB
		}),
	}, nil
}

func NewKafkaWriter(topic string, addr []string, maxMessages int) (*Writer, error) {
	if len(addr) == 0 {
		return nil, ErrNoBrokers
	}
	if topic == "" {
		return nil, ErrTopicRequired
	}
	return &Writer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(addr...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		m: maxMessages,
	}, nil
}

func (w *Writer) AddMessage(key, value []byte) {
	w.mu.Lock()
	w.messages = append(w.messages, &Message{Key: key, Value: value})
	w.mu.Unlock()
}

func (w *Writer) Len() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.messages)
}

func (w *Writer) MaxLen() int {
	return w.m
}

func (w *Writer) clearLocked() {
	w.messages = w.messages[:0]
}

func (w *Writer) WriteMessage(ctx context.Context) error {
	w.mu.Lock()
	m := w.Len()
	if m == 0 {
		return ErrNoMessages
	}
	kafkaMsgs := make([]kafka.Message, m)
	for i, m := range w.messages {
		kafkaMsgs[i] = kafka.Message{
			Key:   m.Key,
			Value: m.Value,
		}
	}
	w.clearLocked()
	w.mu.Unlock()
	if err := w.writer.WriteMessages(ctx, kafkaMsgs...); err != nil {
		return err
	}
	return nil
}

func (r *Reader) ReadMessage(ctx context.Context) (*Message, error) {
	msg, err := r.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return &Message{
		Key:   msg.Key,
		Value: msg.Value,
	}, nil
}

func (w *Writer) Close() error {
	return w.writer.Close()
}

func (r *Reader) Close() error {
	return r.reader.Close()
}
