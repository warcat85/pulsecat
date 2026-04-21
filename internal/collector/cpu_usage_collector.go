package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"pulsecat/internal/metrics"
	"strconv"
	"strings"
	"time"
)

// represents CPU usage data in internal format.
type CPUUsage struct {
	User      float64
	System    float64
	Idle      float64
	Nice      float64
	Iowait    float64
	Irq       float64
	SoftIrq   float64
	Steal     float64
	Guest     float64
	GuestNice float64
}

// a placeholder collector that returns simulated CPU usage data.
type DummyCPUUsageCollector struct{}

// creates a new dummy CPU usage collector.
func NewDummyCPUUsageCollector() Collector {
	return &DummyCPUUsageCollector{}
}

// returns the metric type for CPU usage.
func (c *DummyCPUUsageCollector) Type() metrics.MetricType {
	return metrics.CPUUsage
}

// returns a human-readable name for this collector.
func (c *DummyCPUUsageCollector) Name() string {
	return "dummy_cpu_usage"
}

// returns a simulated CPU usage snapshot.
func (c *DummyCPUUsageCollector) Collect(_ context.Context) (metrics.Sample, error) {
	now := time.Now()
	second := now.Second()
	return &CPUUsage{
		User:   10.5 + float64(second%10),
		System: 5.2 + float64(second%5),
		Idle:   84.3 - float64(second%15),
	}, nil
}

const cpuProcFile = "/proc/stat"

type procData struct {
	values []uint64
	total  uint64
}

// a collector that returns CPU usage data.
type CPUUsageCollector struct {
	prev *procData
}

// creates a new CPU usage collector.
func NewCPUUsageCollector() Collector {
	return &CPUUsageCollector{}
}

// returns the metric type for CPU usage.
func (c *CPUUsageCollector) Type() metrics.MetricType {
	return metrics.CPUUsage
}

// returns a human-readable name for this collector.
func (c *CPUUsageCollector) Name() string {
	return "load_average"
}

// returns a load average snapshot.
func (c *CPUUsageCollector) Collect(_ context.Context) (metrics.Sample, error) {
	data, err := collectData()
	if err != nil {
		// if cannot open, return error
		return nil, fmt.Errorf("unable to collect data: %w", err)
	}
	// no previous results - return 0
	// this will happen always on the first run
	if c.prev == nil {
		c.prev = data
		return &CPUUsage{}, nil
	}
	deltas := make([]uint64, 10)
	for i, value := range data.values {
		deltas[i] = value - c.prev.values[i]
	}
	totalDelta := float64(data.total - c.prev.total)
	fmt.Printf("Data: %.2f (%v)\n", totalDelta, deltas)
	c.prev = data
	return &CPUUsage{
		User:      100 * float64(deltas[0]) / totalDelta,
		System:    100 * float64(deltas[2]) / totalDelta,
		Idle:      100 * float64(deltas[3]) / totalDelta,
		Nice:      100 * float64(deltas[1]) / totalDelta,
		Iowait:    100 * float64(deltas[4]) / totalDelta,
		Irq:       100 * float64(deltas[5]) / totalDelta,
		SoftIrq:   100 * float64(deltas[6]) / totalDelta,
		Steal:     100 * float64(deltas[7]) / totalDelta,
		Guest:     100 * float64(deltas[8]) / totalDelta,
		GuestNice: 100 * float64(deltas[9]) / totalDelta,
	}, nil
}

func collectData() (*procData, error) {
	file, err := os.Open(cpuProcFile)
	if err != nil {
		// if cannot open, return error
		return nil, fmt.Errorf("cannot open %s: %w", cpuProcFile, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// columns are: user nice system idle iowait irq softirq steal guest guest_nice
	// user = user + nice + guest + guest_nice
	// system = system + irq + softirq
	// idle = idle + iowait

	// first is the word 'cpu '
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", cpuProcFile, err)
	}
	words := strings.Fields(line)
	if len(words) != 11 || words[0] != "cpu" {
		return nil, fmt.Errorf("cannot parse %s: %w", line, err)
	}

	values := make([]uint64, 10)
	var total uint64
	// words except first
	for i, word := range words[1:] {
		value, err := strconv.ParseUint(word, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %s: %w", word, err)
		}
		values[i] += value
		total += value
	}

	return &procData{
		values: values,
		total:  total,
	}, nil
}
