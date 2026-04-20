package runner

import (
	"context"
	"log"
	"pulsecat/internal/collector"
	"time"
)

type Runner struct {
	collector collector.Collector
	cancel    context.CancelFunc
	consumer  Consumer
	interval  time.Duration
}

func NewRunner(collector collector.Collector, consumer Consumer, interval time.Duration) *Runner {
	return &Runner{
		collector: collector,
		consumer:  consumer,
		interval:  interval,
	}
}

// Collector returns the collector associated with this runner.
func (c *Runner) Collector() collector.Collector {
	return c.collector
}

// Consumer returns the consumer associated with this runner.
func (c *Runner) Consumer() Consumer {
	return c.consumer
}

// runs the collection loop.
func (c *Runner) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot, err := c.collector.Collect(ctx)
			if err != nil {
				log.Printf("Collector %s failed: %v", c.collector.Name(), err)
				continue
			}
			err = c.consumer.Consume(ctx, snapshot)
			if err != nil {
				log.Printf("Consumer failed: %v", err)
				continue
			}
		}
	}
}

func (c *Runner) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}
