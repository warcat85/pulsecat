package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents a single disk's usage data in internal format.
type DiskUsage struct {
	Filesystem  string
	TotalMb     uint64
	UsedMb      uint64
	UsedPercent float64
	MountPoint  string
}

// represents a collection of disk usage data in internal format.
type DiskUsages struct {
	Disks []*DiskUsage
}

// aplaceholder collector that returns simulated disk usage data.
type DummyDiskUsageCollector struct {
	PeriodicCollector
}

// creates a new dummy disk usage collector.
func NewDummyDiskUsageCollector() *DummyDiskUsageCollector {
	return &DummyDiskUsageCollector{}
}

// returns the metric type for disk usage.
func (c *DummyDiskUsageCollector) Type() metrics.MetricType {
	return metrics.DISK_USAGE
}

// returns a human-readable name for this collector.
func (c *DummyDiskUsageCollector) Name() string {
	return "dummy_disk_usage"
}

// returns a simulated disk usage snapshot.
// The data matches the logic in server.CollectStatistics.
func (c *DummyDiskUsageCollector) Collect(ctx context.Context) (any, error) {
	now := time.Now()
	second := now.Second()
	return &DiskUsages{
		Disks: []*DiskUsage{
			{
				Filesystem:  "/dev/sda1",
				TotalMb:     102400,
				UsedMb:      51200 + uint64(second%100),
				UsedPercent: 50.0 + float64(second%10)*0.1,
				MountPoint:  "/",
			},
		},
	}, nil
}
