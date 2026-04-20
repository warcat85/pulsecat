package average

import (
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
	"testing"
)

func TestCalculateLoadAverage(t *testing.T) {
	tests := []struct {
		name     string
		samples  metrics.Samples
		expected *collector.LoadAverage
	}{
		{
			name:    "empty samples",
			samples: metrics.Samples{},
			expected: &collector.LoadAverage{
				OneMin:     0,
				FiveMin:    0,
				FifteenMin: 0,
			},
		},
		{
			name: "single sample",
			samples: metrics.Samples{
				&collector.LoadAverage{OneMin: 1.0, FiveMin: 2.0, FifteenMin: 3.0},
			},
			expected: &collector.LoadAverage{
				OneMin:     1.0,
				FiveMin:    2.0,
				FifteenMin: 3.0,
			},
		},
		{
			name: "multiple samples",
			samples: metrics.Samples{
				&collector.LoadAverage{OneMin: 1.0, FiveMin: 2.0, FifteenMin: 3.0},
				&collector.LoadAverage{OneMin: 3.0, FiveMin: 4.0, FifteenMin: 5.0},
				&collector.LoadAverage{OneMin: 5.0, FiveMin: 6.0, FifteenMin: 7.0},
			},
			expected: &collector.LoadAverage{
				OneMin:     (1.0 + 3.0 + 5.0) / 3,
				FiveMin:    (2.0 + 4.0 + 6.0) / 3,
				FifteenMin: (3.0 + 5.0 + 7.0) / 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLoadAverage(tt.samples)
			loadAvg, ok := result.(*collector.LoadAverage)
			if !ok {
				t.Fatalf("expected *collector.LoadAverage, got %T", result)
			}

			if loadAvg.OneMin != tt.expected.OneMin {
				t.Errorf("OneMin: got %v, want %v", loadAvg.OneMin, tt.expected.OneMin)
			}
			if loadAvg.FiveMin != tt.expected.FiveMin {
				t.Errorf("FiveMin: got %v, want %v", loadAvg.FiveMin, tt.expected.FiveMin)
			}
			if loadAvg.FifteenMin != tt.expected.FifteenMin {
				t.Errorf("FifteenMin: got %v, want %v", loadAvg.FifteenMin, tt.expected.FifteenMin)
			}
		})
	}
}

func TestCalculateCPUUsage(t *testing.T) {
	tests := []struct {
		name     string
		samples  metrics.Samples
		expected *collector.CPUUsage
	}{
		{
			name:     "empty samples",
			samples:  metrics.Samples{},
			expected: &collector.CPUUsage{User: 0, System: 0, Idle: 0},
		},
		{
			name: "single sample",
			samples: metrics.Samples{
				&collector.CPUUsage{User: 10.0, System: 20.0, Idle: 70.0},
			},
			expected: &collector.CPUUsage{User: 10.0, System: 20.0, Idle: 70.0},
		},
		{
			name: "multiple samples",
			samples: metrics.Samples{
				&collector.CPUUsage{User: 10.0, System: 20.0, Idle: 70.0},
				&collector.CPUUsage{User: 20.0, System: 30.0, Idle: 50.0},
				&collector.CPUUsage{User: 30.0, System: 10.0, Idle: 60.0},
			},
			expected: &collector.CPUUsage{
				User:   (10.0 + 20.0 + 30.0) / 3,
				System: (20.0 + 30.0 + 10.0) / 3,
				Idle:   (70.0 + 50.0 + 60.0) / 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCPUUsage(tt.samples)
			cpuUsage, ok := result.(*collector.CPUUsage)
			if !ok {
				t.Fatalf("expected *collector.CPUUsage, got %T", result)
			}

			if cpuUsage.User != tt.expected.User {
				t.Errorf("User: got %v, want %v", cpuUsage.User, tt.expected.User)
			}
			if cpuUsage.System != tt.expected.System {
				t.Errorf("System: got %v, want %v", cpuUsage.System, tt.expected.System)
			}
			if cpuUsage.Idle != tt.expected.Idle {
				t.Errorf("Idle: got %v, want %v", cpuUsage.Idle, tt.expected.Idle)
			}
		})
	}
}

func TestCalculateNetworkStats(t *testing.T) {
	tests := []struct {
		name     string
		samples  metrics.Samples
		expected *collector.NetworkStats
	}{
		{
			name:     "empty samples",
			samples:  metrics.Samples{},
			expected: &collector.NetworkStats{TotalBytesReceived: 0, TotalBytesSent: 0},
		},
		{
			name: "single sample",
			samples: metrics.Samples{
				&collector.NetworkStats{TotalBytesReceived: 1000, TotalBytesSent: 500},
			},
			expected: &collector.NetworkStats{TotalBytesReceived: 1000, TotalBytesSent: 500},
		},
		{
			name: "multiple samples with integer division",
			samples: metrics.Samples{
				&collector.NetworkStats{TotalBytesReceived: 1000, TotalBytesSent: 500},
				&collector.NetworkStats{TotalBytesReceived: 2000, TotalBytesSent: 1000},
				&collector.NetworkStats{TotalBytesReceived: 3000, TotalBytesSent: 1500},
			},
			expected: &collector.NetworkStats{
				TotalBytesReceived: (1000 + 2000 + 3000) / 3,
				TotalBytesSent:     (500 + 1000 + 1500) / 3,
			},
		},
		{
			name: "division with remainder - integer truncation",
			samples: metrics.Samples{
				&collector.NetworkStats{TotalBytesReceived: 1, TotalBytesSent: 2},
				&collector.NetworkStats{TotalBytesReceived: 2, TotalBytesSent: 3},
			},
			expected: &collector.NetworkStats{
				TotalBytesReceived: (1 + 2) / 2, // 1.5 truncated to 1
				TotalBytesSent:     (2 + 3) / 2, // 2.5 truncated to 2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNetworkStats(tt.samples)
			netStats, ok := result.(*collector.NetworkStats)
			if !ok {
				t.Fatalf("expected *collector.NetworkStats, got %T", result)
			}

			if netStats.TotalBytesReceived != tt.expected.TotalBytesReceived {
				t.Errorf("TotalBytesReceived: got %v, want %v", netStats.TotalBytesReceived, tt.expected.TotalBytesReceived)
			}
			if netStats.TotalBytesSent != tt.expected.TotalBytesSent {
				t.Errorf("TotalBytesSent: got %v, want %v", netStats.TotalBytesSent, tt.expected.TotalBytesSent)
			}
		})
	}
}

func TestCalculateTCPConnectionStates(t *testing.T) {
	tests := []struct {
		name     string
		samples  metrics.Samples
		expected *collector.TCPConnectionStates
	}{
		{
			name:     "empty samples",
			samples:  metrics.Samples{},
			expected: &collector.TCPConnectionStates{Established: 0, Listen: 0},
		},
		{
			name: "single sample",
			samples: metrics.Samples{
				&collector.TCPConnectionStates{Established: 10, Listen: 5},
			},
			expected: &collector.TCPConnectionStates{Established: 10, Listen: 5},
		},
		{
			name: "multiple samples",
			samples: metrics.Samples{
				&collector.TCPConnectionStates{Established: 10, Listen: 5},
				&collector.TCPConnectionStates{Established: 20, Listen: 10},
				&collector.TCPConnectionStates{Established: 30, Listen: 15},
			},
			expected: &collector.TCPConnectionStates{
				Established: (10 + 20 + 30) / 3,
				Listen:      (5 + 10 + 15) / 3,
			},
		},
		{
			name: "division with remainder - integer truncation",
			samples: metrics.Samples{
				&collector.TCPConnectionStates{Established: 1, Listen: 1},
				&collector.TCPConnectionStates{Established: 2, Listen: 2},
			},
			expected: &collector.TCPConnectionStates{
				Established: (1 + 2) / 2, // 1.5 truncated to 1
				Listen:      (1 + 2) / 2, // 1.5 truncated to 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTCPConnectionStates(tt.samples)
			tcpStates, ok := result.(*collector.TCPConnectionStates)
			if !ok {
				t.Fatalf("expected *collector.TCPConnectionStates, got %T", result)
			}

			if tcpStates.Established != tt.expected.Established {
				t.Errorf("Established: got %v, want %v", tcpStates.Established, tt.expected.Established)
			}
			if tcpStates.Listen != tt.expected.Listen {
				t.Errorf("Listen: got %v, want %v", tcpStates.Listen, tt.expected.Listen)
			}
		})
	}
}

func TestCalculateFunctionsWithWrongTypePanic(t *testing.T) {
	// These tests verify that passing wrong sample types causes a panic
	// due to type assertion in the functions
	t.Run("LoadAverage with wrong type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for wrong sample type")
			}
		}()
		samples := metrics.Samples{&collector.CPUUsage{}}
		CalculateLoadAverage(samples)
	})

	t.Run("CPUUsage with wrong type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for wrong sample type")
			}
		}()
		samples := metrics.Samples{&collector.LoadAverage{}}
		CalculateCPUUsage(samples)
	})

	t.Run("NetworkStats with wrong type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for wrong sample type")
			}
		}()
		samples := metrics.Samples{&collector.CPUUsage{}}
		CalculateNetworkStats(samples)
	})

	t.Run("TCPConnectionStates with wrong type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for wrong sample type")
			}
		}()
		samples := metrics.Samples{&collector.NetworkStats{}}
		CalculateTCPConnectionStates(samples)
	})
}
