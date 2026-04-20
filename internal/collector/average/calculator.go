package average

import (
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
)

func CalculateLoadAverage(samples metrics.Samples) metrics.Sample {
	n := float64(len(samples))
	if n == 0 {
		return &collector.LoadAverage{
			OneMin:     0,
			FiveMin:    0,
			FifteenMin: 0,
		}
	}
	var sumOneMin, sumFiveMin, sumFifteenMin float64
	for _, sample := range samples {
		sample := sample.(*collector.LoadAverage)
		sumOneMin += sample.OneMin
		sumFiveMin += sample.FiveMin
		sumFifteenMin += sample.FifteenMin
	}
	return &collector.LoadAverage{
		OneMin:     sumOneMin / n,
		FiveMin:    sumFiveMin / n,
		FifteenMin: sumFifteenMin / n,
	}
}

func CalculateCPUUsage(samples metrics.Samples) metrics.Sample {
	n := float64(len(samples))
	if n == 0 {
		return &collector.CPUUsage{
			User:   0,
			System: 0,
			Idle:   0,
		}
	}
	var sumUser, sumSystem, sumIdle float64
	for _, sample := range samples {
		sample := sample.(*collector.CPUUsage)
		sumUser += sample.User
		sumSystem += sample.System
		sumIdle += sample.Idle
	}
	return &collector.CPUUsage{
		User:   sumUser / n,
		System: sumSystem / n,
		Idle:   sumIdle / n,
	}
}

func CalculateNetworkStats(samples metrics.Samples) metrics.Sample {
	n := uint64(len(samples))
	if n == 0 {
		return &collector.NetworkStats{
			TotalBytesReceived: 0,
			TotalBytesSent:     0,
		}
	}
	var sumReceive, sumTransmit uint64
	for _, sample := range samples {
		sample := sample.(*collector.NetworkStats)
		sumReceive += sample.TotalBytesReceived
		sumTransmit += sample.TotalBytesSent
	}
	return &collector.NetworkStats{
		TotalBytesReceived: sumReceive / n,
		TotalBytesSent:     sumTransmit / n,
	}
}

func CalculateTCPConnectionStates(samples metrics.Samples) metrics.Sample {
	n := uint32(len(samples)) //nolint:gosec // size of array cannot be out of range
	if n == 0 {
		return &collector.TCPConnectionStates{
			Established: 0,
			Listen:      0,
		}
	}
	var sumEstablished, sumListen uint32
	for _, sample := range samples {
		sample := sample.(*collector.TCPConnectionStates)
		sumEstablished += sample.Established
		sumListen += sample.Listen
	}

	return &collector.TCPConnectionStates{
		Established: sumEstablished / n,
		Listen:      sumListen / n,
	}
}
