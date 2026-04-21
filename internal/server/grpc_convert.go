package server

import (
	"fmt"
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
	v1 "pulsecat/pkg/api/v1"
	"time"
)

var converters = metrics.MetricMap[Converter]{
	metrics.Meow:                MeowConverter{},
	metrics.LoadAverage:         LoadAverageConverter{},
	metrics.CPUUsage:            CPUUsageConverter{},
	metrics.DiskUsage:           DiskUsageConverter{},
	metrics.NetworkStats:        NetworkStatsConverter{},
	metrics.TopTalkers:          TopTalkersConverter{},
	metrics.ListeningSockets:    ListeningSocketsConverter{},
	metrics.TCPConnectionStates: TCPConnectionStatesConverter{},
}

type Converter interface {
	Name() string
	Convert(metrics.Sample) (*v1.MetricPulse, error)
}

// MeowConverter is a converter for Meow.
type MeowConverter struct{}

func (MeowConverter) Name() string { return "meow" }

// Convert implements the Converter interface for MeowConverter.
func (c MeowConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.Meow)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_Meow{
		Meow: ConvertMeow(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// LoadAverageConverter is a converter for load average.
type LoadAverageConverter struct{}

func (LoadAverageConverter) Name() string { return "load average" }

// Convert implements the Converter interface for LoadAverageConverter.
func (c LoadAverageConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.LoadAverage)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_LoadAverage{
		LoadAverage: ConvertLoadAverage(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// CPUUsageConverter is a converter for CPU usage.
type CPUUsageConverter struct{}

func (CPUUsageConverter) Name() string { return "CPU usage" }

// Convert implements the Converter interface for CPUUsageConverter.
func (c CPUUsageConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.CPUUsage)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_CpuUsage{
		CpuUsage: ConvertCPUUsage(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// DiskUsageConverter is a converter for disk usage.
type DiskUsageConverter struct{}

func (DiskUsageConverter) Name() string { return "disk usage" }

// Convert implements the Converter interface for DiskUsageConverter.
func (c DiskUsageConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.DiskUsages)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_DiskUsage{
		DiskUsage: ConvertDiskUsagesToProto(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// NetworkStatsConverter is a converter for network stats.
type NetworkStatsConverter struct{}

func (NetworkStatsConverter) Name() string { return "network stats" }

// Convert implements the Converter interface for NetworkStatsConverter.
func (c NetworkStatsConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.NetworkStats)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_NetworkStats{
		NetworkStats: ConvertNetworkStats(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// TopTalkersConverter is a converter for top talkers.
type TopTalkersConverter struct{}

func (TopTalkersConverter) Name() string { return "top talkers" }

// Convert implements the Converter interface for TopTalkersConverter.
func (c TopTalkersConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.NetworkTalkers)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_TopTalkers{
		TopTalkers: ConvertNetworkTalkers(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// ListeningSocketsConverter is a converter for listening sockets.
type ListeningSocketsConverter struct{}

func (ListeningSocketsConverter) Name() string { return "listening sockets" }

// Convert implements the Converter interface for ListeningSocketsConverter.
func (c ListeningSocketsConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.ListeningSockets)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_ListeningSockets{
		ListeningSockets: ConvertListeningSockets(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// TCPConnectionStatesConverter is a converter for TCP connection states.
type TCPConnectionStatesConverter struct{}

func (TCPConnectionStatesConverter) Name() string { return "TCP connection states" }

// Convert implements the Converter interface for TCPConnectionStatesConverter.
func (c TCPConnectionStatesConverter) Convert(sample metrics.Sample) (*v1.MetricPulse, error) {
	data, ok := sample.(*collector.TCPConnectionStates)
	if !ok {
		return nil, fmt.Errorf("unexpected type from %s collector: %T", c.Name(), sample)
	}
	metric := &v1.MetricPulse_TcpConnectionStates{
		TcpConnectionStates: ConvertTCPConnectionStates(data),
	}
	return &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
		Metric:    metric,
	}, nil
}

// converts internal Meow to protobuf Meow.
func ConvertMeow(in *collector.Meow) *v1.Meow {
	if in == nil {
		return nil
	}
	return &v1.Meow{
		Message: in.Message,
	}
}

// converts internal LoadAverage to protobuf LoadAverage.
func ConvertLoadAverage(in *collector.LoadAverage) *v1.LoadAverage {
	if in == nil {
		return nil
	}
	return &v1.LoadAverage{
		OneMin:     in.OneMin,
		FiveMin:    in.FiveMin,
		FifteenMin: in.FifteenMin,
	}
}

// converts internal CPUUsage to protobuf CPUUsage.
func ConvertCPUUsage(in *collector.CPUUsage) *v1.CPUUsage {
	if in == nil {
		return nil
	}
	return &v1.CPUUsage{
		User:      in.User,
		System:    in.System,
		Idle:      in.Idle,
		Nice:      in.Nice,
		Iowait:    in.Iowait,
		Irq:       in.Irq,
		SoftIrq:   in.SoftIrq,
		Steal:     in.Steal,
		Guest:     in.Guest,
		GuestNice: in.GuestNice,
	}
}

// converts internal DiskUsage to protobuf DiskUsage.
func ConvertDiskUsage(in *collector.DiskUsage) *v1.DiskUsage {
	if in == nil {
		return nil
	}
	return &v1.DiskUsage{
		Filesystem:  in.Filesystem,
		TotalMb:     in.TotalMb,
		UsedMb:      in.UsedMb,
		UsedPercent: in.UsedPercent,
		MountPoint:  in.MountPoint,
		// AvailableMb, TotalInodes, UsedInodes, UsedInodesPercent left zero
	}
}

// converts internal DiskUsages to protobuf slice of DiskUsage.
func ConvertDiskUsages(in *collector.DiskUsages) []*v1.DiskUsage {
	if in == nil || in.Disks == nil {
		return nil
	}
	out := make([]*v1.DiskUsage, len(in.Disks))
	for i, d := range in.Disks {
		out[i] = ConvertDiskUsage(d)
	}
	return out
}

// converts internal DiskUsages to protobuf DiskUsages.
func ConvertDiskUsagesToProto(in *collector.DiskUsages) *v1.DiskUsages {
	if in == nil {
		return nil
	}
	return &v1.DiskUsages{
		Disks: ConvertDiskUsages(in),
	}
}

// converts internal NetworkStats to protobuf NetworkStats.
func ConvertNetworkStats(in *collector.NetworkStats) *v1.NetworkStats {
	if in == nil {
		return nil
	}
	return &v1.NetworkStats{
		TotalBytesReceived: in.TotalBytesReceived,
		TotalBytesSent:     in.TotalBytesSent,
		// PacketsReceived, PacketsSent, ErrorsReceived, ErrorsSent, DropsReceived, DropsSent left zero
	}
}

// converts internal ProtocolTalker to protobuf ProtocolTalker.
func ConvertProtocolTalker(in *collector.ProtocolTalker) *v1.ProtocolTalker {
	if in == nil {
		return nil
	}
	return &v1.ProtocolTalker{
		Protocol: in.Protocol,
		Port:     in.Port,
	}
}

// converts internal NetworkTalker to protobuf NetworkTalker.
func ConvertNetworkTalker(in *collector.NetworkTalker) *v1.NetworkTalker {
	if in == nil {
		return nil
	}
	if in.Protocol != nil {
		return &v1.NetworkTalker{
			Identifier: &v1.NetworkTalker_Protocol{
				Protocol: ConvertProtocolTalker(in.Protocol),
			},
			BytesPerSecond: in.BytesPerSecond,
			Percentage:     in.Percentage,
		}
	}
	// Protocol is nil, set Identifier to nil.
	return &v1.NetworkTalker{
		Identifier:     nil,
		BytesPerSecond: in.BytesPerSecond,
		Percentage:     in.Percentage,
	}
}

// converts internal NetworkTalkers to protobuf NetworkTalkers.
func ConvertNetworkTalkers(in *collector.NetworkTalkers) *v1.NetworkTalkers {
	if in == nil || in.Talkers == nil {
		return nil
	}
	out := &v1.NetworkTalkers{
		Talkers: make([]*v1.NetworkTalker, len(in.Talkers)),
	}
	for i, t := range in.Talkers {
		out.Talkers[i] = ConvertNetworkTalker(t)
	}
	return out
}

// converts internal ListeningSocket to protobuf ListeningSocket.
func ConvertListeningSocket(in *collector.ListeningSocket) *v1.ListeningSocket {
	if in == nil {
		return nil
	}
	return &v1.ListeningSocket{
		Command:  in.Command,
		Pid:      in.Pid,
		User:     in.User,
		Protocol: in.Protocol,
		Port:     in.Port,
		Address:  in.Address,
	}
}

// converts internal ListeningSockets to protobuf ListeningSockets.
func ConvertListeningSockets(in *collector.ListeningSockets) *v1.ListeningSockets {
	if in == nil || in.Sockets == nil {
		return nil
	}
	out := &v1.ListeningSockets{
		Sockets: make([]*v1.ListeningSocket, len(in.Sockets)),
	}
	for i, s := range in.Sockets {
		out.Sockets[i] = ConvertListeningSocket(s)
	}
	return out
}

// converts internal TCPConnectionStates to protobuf TcpConnectionStates.
func ConvertTCPConnectionStates(in *collector.TCPConnectionStates) *v1.TCPConnectionStates {
	if in == nil {
		return nil
	}
	return &v1.TCPConnectionStates{
		Established: in.Established,
		Listen:      in.Listen,
		// SynSent, SynRecv, FinWait1, FinWait2, TimeWait, Close, CloseWait, LastAck, Closing left zero
	}
}

// converts a protobuf MetricType to internal MetricType.
// Returns the internal metric type and true if the conversion is successful.
// For unsupported metric types (UNSPECIFIED) returns false.
func ConvertMetricTypeFromProto(protoType v1.MetricType) (metrics.MetricType, bool) {
	switch protoType {
	case v1.MetricType_METRIC_TYPE_MEOW:
		return metrics.Meow, true
	case v1.MetricType_METRIC_TYPE_LOAD_AVERAGE:
		return metrics.LoadAverage, true
	case v1.MetricType_METRIC_TYPE_CPU_USAGE:
		return metrics.CPUUsage, true
	case v1.MetricType_METRIC_TYPE_DISK_USAGE:
		return metrics.DiskUsage, true
	case v1.MetricType_METRIC_TYPE_NETWORK_STATS:
		return metrics.NetworkStats, true
	case v1.MetricType_METRIC_TYPE_TOP_TALKERS:
		return metrics.TopTalkers, true
	case v1.MetricType_METRIC_TYPE_LISTENING_SOCKETS:
		return metrics.ListeningSockets, true
	case v1.MetricType_METRIC_TYPE_TCP_CONNECTION_STATES:
		return metrics.TCPConnectionStates, true
	case v1.MetricType_METRIC_TYPE_UNSPECIFIED:
		return 0, false
	default:
		return 0, false
	}
}
