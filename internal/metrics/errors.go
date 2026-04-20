package metrics

import (
	"fmt"
)

type UnsupportedMetricType struct {
	metricType MetricType
}

func ErrUnsupportedMetricType(metricType MetricType) error {
	return &UnsupportedMetricType{metricType: metricType}
}

func (e *UnsupportedMetricType) Error() string {
	return fmt.Sprintf("unsupported metric type: %d", e.metricType)
}

type CollectorDisabled struct {
	metricType MetricType
}

func ErrCollectorDisabled(metricType MetricType) error {
	return &CollectorDisabled{metricType: metricType}
}

func (e *CollectorDisabled) Error() string {
	return fmt.Sprintf("collector is disabled for metric type: %v", e.metricType)
}
