package metrics

import (
	"fmt"
)

type MetricType int

const (
	LoadAverage MetricType = iota
	CPUUsage
	DiskUsage
	NetworkStats
	TopTalkers
	ListeningSockets
	TCPConnectionStates
	Meow
)

var metricNames = MetricMap[string]{
	LoadAverage:         "load_average",
	CPUUsage:            "cpu_usage",
	DiskUsage:           "disk_usage",
	NetworkStats:        "network_stats",
	TopTalkers:          "top_talkers",
	ListeningSockets:    "listening_sockets",
	TCPConnectionStates: "tcp_connection_states",
	Meow:                "meow",
}

func (t MetricType) String() string {
	if s, ok := metricNames[t]; ok {
		return s
	}
	panic(fmt.Sprintf("unknown metric type: %d", t))
}

type (
	MetricMap[T any] map[MetricType]T
)

// a single metric sample.
type Sample any

type Samples []Sample

type AverageCalculator func(Samples) Sample
