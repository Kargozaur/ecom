package processor

import (
	"context"
	"errors"
	"log"
	"order/repo"
	"pkg/broker"
	"pkg/envreader"
	orevents "proto/out/events/v1"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

type Notifier interface {
	Notify()
}

type Processor struct {
	wr   *broker.Writer
	rd   *broker.Reader
	th   chan struct{}
	done chan struct{}
	repo *repo.Repo
}

func NewProcessor(pool *pgxpool.Pool) *Processor {
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
		repo: repo.NewRepo(pool),
	}
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	defer close(p.done)
	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			for {
				if err := p.write(shutdownCtx); err != nil {
					if errors.Is(err, broker.ErrNoMessages) {
						log.Println("no messages to send")
						return
					}
					log.Printf("failed to send message: %s\n", err.Error())
					return
				}
			}
		case <-ticker.C:
			p.handleWrite(ctx)
		case <-p.th:
			p.handleWrite(ctx)
			ticker.Reset(time.Second * 30)
		}
	}
}

func (p *Processor) handleWrite(ctx context.Context) {
	if err := p.write(ctx); err != nil {
		if errors.Is(err, broker.ErrNoMessages) {
			log.Println("no messages to send")
			return
		}
		log.Printf("failed to send message: %s\n", err.Error())
	}
}

func (p *Processor) append(key, value []byte) {
	p.wr.AddMessage(key, value)
	if p.wr.Len() >= p.wr.MaxLen() {
		select {
		case p.th <- struct{}{}:
		default:
		}
	}
}

func (p *Processor) fillMessages(ctx context.Context) error {
	r := &orevents.Events{
		Events: make([]*orevents.Event, 0, 80),
	}
	err := p.repo.WithinTx(ctx, func(c context.Context) error {
		rows, key, err := p.repo.UpdateEvent(c, 80)
		if err != nil {
			return err
		}
		for _, row := range rows {
			r.Events = append(r.Events, &orevents.Event{
				EventId: row.ID.String(),
				Status:  row.EventKey.String(),
			})
		}
		body, err := proto.Marshal(r)
		if err != nil {
			return err
		}
		p.append(key[:], body)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Processor) write(ctx context.Context) error {
	if err := p.fillMessages(ctx); err != nil {
		return err
	}
	if err := p.wr.WriteMessage(ctx); err != nil {
		return err
	}
	return nil
}

func (p *Processor) Close() error {
	rdErr := p.rd.Close()
	<-p.done
	wrErr := p.wr.Close()
	return errors.Join(wrErr, rdErr)
}

func (p *Processor) Notify() {
	select {
	case p.th <- struct{}{}:
	default:
	}
}
