package processor

import (
	"context"
	"errors"
	"log"
	"pkg/broker"
	"pkg/envreader"
	"strings"
	"time"
)

type Processor struct {
	wr   *broker.Writer
	rd   *broker.Reader
	th   chan struct{}
	done chan struct{}
}

func NewProcessor() *Processor {
	readerTopic := envreader.Read("ORDER_READER", "payments")
	brokers := envreader.Read("KAFKA_BROKERS", "")
	if brokers == "" {
		log.Fatalf("KAFKA_BROKERS environment variable is not set\n")
	}
	sl := strings.Split(brokers, ",")
	rd, err := broker.NewKafkaReader(readerTopic, sl, 0)
	if err != nil {
		log.Fatalf("failed to create kafka reader: %s\n", err.Error())
	}
	writerTopic := envreader.Read("ORDER_WRITER", "orders")
	wr, err := broker.NewKafkaWriter(writerTopic, sl, 100)
	if err != nil {
		log.Fatalf("failed to create kafka writer: %s\n", err.Error())
	}
	return &Processor{
		rd:   rd,
		wr:   wr,
		th:   make(chan struct{}, 1),
		done: make(chan struct{}),
	}
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	defer close(p.done)
	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := p.write(shutdownCtx); err != nil {
				if errors.Is(err, broker.ErrNoMessages) {
					log.Println("no messages to send")
					return
				}
				log.Printf("failed to send message: %s\n", err.Error())
			}
			return
		case <-ticker.C:
			if err := p.write(ctx); err != nil {
				if errors.Is(err, broker.ErrNoMessages) {
					log.Println("no messages to send")
					continue
				}
				log.Printf("failed to send message: %s\n", err.Error())
			}
		case <-p.th:
			if err := p.write(ctx); err != nil {
				log.Printf("failed to send message: %s\n", err.Error())
			}
		}
	}
}

func (p *Processor) write(ctx context.Context) error {
	return p.wr.WriteMessage(ctx)
}

func (p *Processor) Append(key, value []byte) {
	p.wr.AddMessage(key, value)
	if p.wr.Len() >= p.wr.MaxLen() {
		select {
		case p.th <- struct{}{}:
		default:
		}
	}
}

func (p *Processor) Close() error {
	rdErr := p.rd.Close()
	<-p.done
	wrErr := p.wr.Close()
	return errors.Join(wrErr, rdErr)
}
