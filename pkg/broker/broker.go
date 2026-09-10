package broker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	Key   []byte
	Value []byte
}

type Writer struct {
	writer       *kafka.Writer
	messages     []*Message
	mu           sync.RWMutex
	maxMessages  int32
	currMessages atomic.Int32
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

func NewKafkaWriter(topic string, addr []string, maxMessages int32) (*Writer, error) {
	if len(addr) == 0 {
		return nil, ErrNoBrokers
	}
	if topic == "" {
		return nil, ErrTopicRequired
	}
	return &Writer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(addr...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: time.Second * 5,
			ReadTimeout:  time.Second * 5,
			MaxAttempts:  5,
		},
		messages:    make([]*Message, maxMessages),
		maxMessages: maxMessages,
	}, nil
}

func (w *Writer) Increase() {
	w.currMessages.Add(1)
}

func (w *Writer) Len() int32 {
	return w.currMessages.Load()
}

func (w *Writer) AddMessage(key, value []byte) {
	w.mu.Lock()
	w.messages = append(w.messages, &Message{Key: key, Value: value})
	w.mu.Unlock()
	w.Increase()
}

func (w *Writer) reduce(delta int32) {
	w.currMessages.Add(-delta)
}

func (w *Writer) clearLocked() {
	w.messages = w.messages[:0]
}

func (w *Writer) MaxLen() int32 {
	return w.maxMessages
}

func (w *Writer) WriteMessage(ctx context.Context) error {
	if w.Len() == 0 {
		return ErrNoMessages
	}

	w.mu.Lock()
	n := len(w.messages)
	kafkaMsgs := make([]kafka.Message, n)
	for i, m := range w.messages {
		kafkaMsgs[i] = kafka.Message{Key: m.Key, Value: m.Value}
	}
	w.mu.Unlock()

	if err := w.writer.WriteMessages(ctx, kafkaMsgs...); err != nil {
		return err
	}

	w.mu.Lock()
	w.messages = w.messages[n:]
	w.mu.Unlock()
	w.reduce(int32(n))
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
