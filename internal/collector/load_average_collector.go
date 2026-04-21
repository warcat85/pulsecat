package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"pulsecat/internal/metrics"
	"strconv"
	"time"
)

// represents load average data in internal format.
type LoadAverage struct {
	OneMin     float64
	FiveMin    float64
	FifteenMin float64
}

// a placeholder collector that returns simulated load average data.
type DummyLoadAverageCollector struct{}

// creates a new dummy load average collector.
func NewDummyLoadAverageCollector() Collector {
	return &DummyLoadAverageCollector{}
}

// returns the metric type for load average.
func (c *DummyLoadAverageCollector) Type() metrics.MetricType {
	return metrics.LoadAverage
}

// returns a human-readable name for this collector.
func (c *DummyLoadAverageCollector) Name() string {
	return "dummy_load_average"
}

// returns a simulated load average snapshot.
func (c *DummyLoadAverageCollector) Collect(_ context.Context) (metrics.Sample, error) {
	now := time.Now()
	second := now.Second()
	baseLoad := 0.1 + float64(second%30)*0.01
	return &LoadAverage{
		OneMin:     baseLoad,
		FiveMin:    baseLoad * 0.9,
		FifteenMin: baseLoad * 0.8,
	}, nil
}

const loadProcFile = "/proc/loadavg"

// a collector that returns load average data.
type LoadAverageCollector struct{}

// creates a new load average collector.
func NewLoadAverageCollector() Collector {
	return &LoadAverageCollector{}
}

// returns the metric type for load average.
func (c *LoadAverageCollector) Type() metrics.MetricType {
	return metrics.LoadAverage
}

// returns a human-readable name for this collector.
func (c *LoadAverageCollector) Name() string {
	return "load_average"
}

// returns a load average snapshot.
func (c *LoadAverageCollector) Collect(_ context.Context) (metrics.Sample, error) {
	// open the procfile
	loadavg, err := os.Open(loadProcFile)
	if err != nil {
		// if cannot open, return error
		return nil, fmt.Errorf("cannot open %s: %w", loadProcFile, err)
	}
	defer loadavg.Close()

	reader := bufio.NewReader(loadavg)
	average := make([]float64, 3)
	var word string
	for i := 0; i < 3; i++ {
		for len(word) <= 1 {
			word, err = reader.ReadString(' ')
			if err != nil {
				return nil, fmt.Errorf("cannot read %s: %w", loadProcFile, err)
			}
		}
		word = word[:len(word)-1]
		// reading without the last character (it is space)
		average[i], err = strconv.ParseFloat(word, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %s: %w", word, err)
		}
	}
	fmt.Printf("Averages: %v\n", average)
	return &LoadAverage{
		OneMin:     average[0],
		FiveMin:    average[1],
		FifteenMin: average[2],
	}, nil
}
