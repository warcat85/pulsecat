package metrics

import (
	"fmt"
	"time"
)

type MetricType int

const (
	LOAD_AVERAGE MetricType = iota
	CPU_USAGE
	DISK_USAGE
	NETWORK_STATS
	TOP_TALKERS
	LISTENING_SOCKETS
	TCP_CONNECTION_STATES
	MEOW
)

var metricNames = MetricMap[string]{
	LOAD_AVERAGE:          "load_average",
	CPU_USAGE:             "cpu_usage",
	DISK_USAGE:            "disk_usage",
	NETWORK_STATS:         "network_stats",
	TOP_TALKERS:           "top_talkers",
	LISTENING_SOCKETS:     "listening_sockets",
	TCP_CONNECTION_STATES: "tcp_connection_states",
	MEOW:                  "meow",
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

// a single metric snapshot with its timestamp.
type Sample struct {
	Timestamp time.Time
	Data      any // concrete type depends on metric type
}
