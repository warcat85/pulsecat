package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	v1 "pulsecat/pkg/api/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("server", "localhost:25225", "PulseCat server address (host:port)")
	startDelay = flag.Uint("delay", 0, "Start delay (in seconds)")
	frequency  = flag.Uint("frequency", 1, "Frequency between snapshots (in seconds)")
	metricType = flag.String("metric", "load", "Metric type to subscribe to (load,cpu,disk,network,talkers,sockets,tcp,meow)")
	duration   = flag.Uint("duration", 0, "Duration in seconds to run (0 = infinite)")
	verbose    = flag.Bool("verbose", false, "Enable verbose output")
	version    = flag.Bool("version", false, "Print version and exit")
)

// information set by build flags
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

func parseMetricType(input string) v1.MetricType {
	input = strings.TrimSpace(strings.ToLower(input))
	switch input {
	case "load":
		return v1.MetricType_METRIC_TYPE_LOAD_AVERAGE
	case "cpu":
		return v1.MetricType_METRIC_TYPE_CPU_USAGE
	case "disk":
		return v1.MetricType_METRIC_TYPE_DISK_USAGE
	case "network":
		return v1.MetricType_METRIC_TYPE_NETWORK_STATS
	case "talkers":
		return v1.MetricType_METRIC_TYPE_TOP_TALKERS
	case "sockets":
		return v1.MetricType_METRIC_TYPE_LISTENING_SOCKETS
	case "tcp":
		return v1.MetricType_METRIC_TYPE_TCP_CONNECTION_STATES
	case "meow":
		return v1.MetricType_METRIC_TYPE_MEOW
	default:
		log.Printf("Warning: unknown metric type '%s', defaulting to load", input)
		return v1.MetricType_METRIC_TYPE_LOAD_AVERAGE
	}
}

func printMetricPulse(pulse *v1.MetricPulse) {
	fmt.Printf("\n=== Metric Pulse at %s ===\n", time.Unix(pulse.Timestamp, 0).Format(time.RFC3339))

	switch metric := pulse.Metric.(type) {
	case *v1.MetricPulse_Meow:
		fmt.Printf("Meow: %s\n", metric.Meow.Message)
	case *v1.MetricPulse_LoadAverage:
		la := metric.LoadAverage
		fmt.Printf("Load Average: 1m=%.2f, 5m=%.2f, 15m=%.2f\n", la.OneMin, la.FiveMin, la.FifteenMin)
	case *v1.MetricPulse_CpuUsage:
		cpu := metric.CpuUsage
		fmt.Printf("CPU Usage: User=%.1f%%, System=%.1f%%, Idle=%.1f%%, Nice=%.1f%%, IOWait=%.1f%%\n",
			cpu.User, cpu.System, cpu.Idle, cpu.Nice, cpu.Iowait)
	case *v1.MetricPulse_DiskUsage:
		disks := metric.DiskUsage.Disks
		fmt.Printf("Disk Usage (%d filesystems):\n", len(disks))
		for _, du := range disks {
			fmt.Printf("  %s: %.1f%% used (%d MB used / %d MB total)\n",
				du.MountPoint, du.UsedPercent, du.UsedMb, du.TotalMb)
		}
	case *v1.MetricPulse_NetworkStats:
		ns := metric.NetworkStats
		fmt.Printf("Network: RX=%d bytes, TX=%d bytes\n", ns.TotalBytesReceived, ns.TotalBytesSent)
	case *v1.MetricPulse_TopTalkers:
		talkers := metric.TopTalkers.Talkers
		fmt.Printf("Top Talkers (%d):\n", len(talkers))
		for i, tt := range talkers {
			if i >= 3 {
				break
			}
			fmt.Printf("  %d. %d bps (%.1f%%)\n", i+1, tt.BytesPerSecond, tt.Percentage)
		}
	case *v1.MetricPulse_ListeningSockets:
		sockets := metric.ListeningSockets.Sockets
		fmt.Printf("Listening Sockets (%d):\n", len(sockets))
		for i, ls := range sockets {
			if i >= 3 {
				break
			}
			fmt.Printf("  %d. %s:%d (%s)\n", i+1, ls.Address, ls.Port, ls.Protocol)
		}
	case *v1.MetricPulse_TcpConnectionStates:
		tcp := metric.TcpConnectionStates
		fmt.Printf("TCP Connections: ESTABLISHED=%d, LISTEN=%d, TIME_WAIT=%d\n",
			tcp.Established, tcp.Listen, tcp.TimeWait)
	default:
		fmt.Printf("Unknown metric type\n")
	}
}

func runSubscription(ctx context.Context, client v1.PulseCatClient) error {
	metric := parseMetricType(*metricType)

	req := &v1.SubscribeRequest{
		StartDelay: uint32(*startDelay),
		Frequency:  uint32(*frequency),
		MetricType: metric,
	}

	if *verbose {
		log.Printf("Connecting to %s", *serverAddr)
		log.Printf("Request: delay=%ds, frequency=%ds, metric_type=%v", req.StartDelay, req.Frequency, metric)
	}

	// Start subscription
	stream, err := client.Subscribe(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}

	log.Printf("Subscribed to PulseCat server. Receiving %s metric every %d seconds...", *metricType, *frequency)
	if *duration > 0 {
		log.Printf("Will run for %d seconds", *duration)
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Set up duration timer if specified
	var durationTimer *time.Timer
	var timerCh <-chan time.Time
	if *duration > 0 {
		durationTimer = time.NewTimer(time.Duration(*duration) * time.Second)
		defer durationTimer.Stop()
		timerCh = durationTimer.C
	}

	snapshotCount := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sigCh:
			log.Printf("Interrupt received, shutting down...")
			return nil
		case <-timerCh:
			log.Printf("Duration reached, shutting down...")
			return nil
		default:
			// Receive next metric pulse
			pulse, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Server closed stream")
				return nil
			}
			if err != nil {
				return fmt.Errorf("error receiving metric pulse: %v", err)
			}

			snapshotCount++
			printMetricPulse(pulse)

			if *verbose {
				log.Printf("Received metric pulse #%d", snapshotCount)
			}
		}
	}
}

func runClient(ctx context.Context) error {
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := v1.NewPulseCatClient(conn)

	return runSubscription(ctx, client)
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("PulseKitten version %s (built %s, commit %s)\n", Version, BuildTime, CommitHash)
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := runClient(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
