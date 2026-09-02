package processor

import (
	"context"
	"log"
	"pkg/broker"
	"time"
)

type Processor struct {
	wr *broker.Writer
	rd *broker.Reader
}

func NewProcessor(wr *broker.Writer, rd *broker.Reader) *Processor {
	return &Processor{
		wr: wr,
		rd: rd,
	}
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := p.wr.WriteMessage(shutdownCtx); err != nil {
				log.Println(err.Error())
			}
			return
		case <-ticker.C:
			if err := p.wr.WriteMessage(ctx); err != nil {
				log.Println(err.Error())
			}
		}
	}
}
