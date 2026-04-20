package average

import (
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
)

func CalculateLoadAverage(samples metrics.Samples) metrics.Sample {
	n := float64(len(samples))
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
