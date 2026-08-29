package broker

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	writer *kafka.Writer
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

func NewKafkaWriter(topic string, addr []string) (*Writer, error) {
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
	}, nil
}

func (w *Writer) WriteMessage(ctx context.Context, key, value []byte) error {
	return w.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (r *Reader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return r.reader.ReadMessage(ctx)
}

func (w *Writer) Close() error {
	return w.writer.Close()
}

func (r *Reader) Close() error {
	return r.reader.Close()
}
