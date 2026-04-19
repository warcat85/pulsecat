package collector

import (
	"context"
	"pulsecat/internal/metrics"
)

type Consumer interface {
	Consume(ctx context.Context, sample metrics.Sample) error
}
