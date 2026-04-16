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
	startDelay = flag.Uint("delay", 0, "Start delay in seconds (M parameter)")
	frequency  = flag.Uint("frequency", 1, "Frequency in seconds between snapshots (N parameter)")
	statTypes  = flag.String("stats", "", "Comma-separated list of stat types to filter (load,cpu,disk,network,talkers,sockets,tcp)")
	duration   = flag.Uint("duration", 0, "Duration in seconds to run (0 = infinite)")
	verbose    = flag.Bool("verbose", false, "Enable verbose output")
	version    = flag.Bool("version", false, "Print version and exit")
)

// Version information set by build flags
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

func parseStatTypes(input string) []v1.StatType {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var stats []v1.StatType
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		switch part {
		case "load":
			stats = append(stats, v1.StatType_STAT_TYPE_LOAD_AVERAGE)
		case "cpu":
			stats = append(stats, v1.StatType_STAT_TYPE_CPU_USAGE)
		case "disk":
			stats = append(stats, v1.StatType_STAT_TYPE_DISK_USAGE)
		case "network":
			stats = append(stats, v1.StatType_STAT_TYPE_NETWORK_STATS)
		case "talkers":
			stats = append(stats, v1.StatType_STAT_TYPE_TOP_TALKERS)
		case "sockets":
			stats = append(stats, v1.StatType_STAT_TYPE_LISTENING_SOCKETS)
		case "tcp":
			stats = append(stats, v1.StatType_STAT_TYPE_TCP_CONNECTION_STATES)
		default:
			log.Printf("Warning: unknown stat type '%s', ignoring", part)
		}
	}
	return stats
}

func printStats(stats *v1.SystemStats) {
	fmt.Printf("\n=== System Stats at %s ===\n", time.Unix(stats.Timestamp, 0).Format(time.RFC3339))

	if stats.LoadAverage != nil {
		la := stats.LoadAverage
		fmt.Printf("Load Average: 1m=%.2f, 5m=%.2f, 15m=%.2f\n", la.OneMin, la.FiveMin, la.FifteenMin)
	}

	if stats.CpuUsage != nil {
		cpu := stats.CpuUsage
		fmt.Printf("CPU Usage: User=%.1f%%, System=%.1f%%, Idle=%.1f%%, Nice=%.1f%%, IOWait=%.1f%%\n",
			cpu.User, cpu.System, cpu.Idle, cpu.Nice, cpu.Iowait)
	}

	if len(stats.DiskUsage) > 0 {
		fmt.Printf("Disk Usage (%d filesystems):\n", len(stats.DiskUsage))
		for i, du := range stats.DiskUsage {
			if i >= 3 && !*verbose { // Limit output unless verbose
				fmt.Printf("  ... and %d more\n", len(stats.DiskUsage)-i)
				break
			}
			fmt.Printf("  %s: %.1f%% used (%d MB used / %d MB total)\n",
				du.MountPoint, du.UsedPercent, du.UsedMb, du.TotalMb)
		}
	}

	if stats.NetworkStats != nil {
		net := stats.NetworkStats
		fmt.Printf("Network: RX=%d bytes, TX=%d bytes\n", net.TotalBytesReceived, net.TotalBytesSent)
	}

	if len(stats.TopTalkers) > 0 {
		fmt.Printf("Top Talkers (%d):\n", len(stats.TopTalkers))
		for i, tt := range stats.TopTalkers {
			if i >= 3 && !*verbose {
				fmt.Printf("  ... and %d more\n", len(stats.TopTalkers)-i)
				break
			}
			fmt.Printf("  %.1f%% at %d bytes/sec\n", tt.Percentage, tt.BytesPerSecond)
		}
	}

	if len(stats.ListeningSockets) > 0 {
		fmt.Printf("Listening Sockets (%d):\n", len(stats.ListeningSockets))
		for i, ls := range stats.ListeningSockets {
			if i >= 3 && !*verbose {
				fmt.Printf("  ... and %d more\n", len(stats.ListeningSockets)-i)
				break
			}
			fmt.Printf("  %s:%d (%s) - %s (PID %d)\n",
				ls.Address, ls.Port, ls.Protocol, ls.Command, ls.Pid)
		}
	}

	if stats.TcpConnectionStates != nil {
		tcp := stats.TcpConnectionStates
		fmt.Printf("TCP Connections: ESTABLISHED=%d, LISTEN=%d, TIME_WAIT=%d\n",
			tcp.Established, tcp.Listen, tcp.TimeWait)
	}
}

func runClient(ctx context.Context) error {
	// Parse stat types
	stats := parseStatTypes(*statTypes)

	// Set up connection
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := v1.NewPulseCatClient(conn)

	// Create subscription request
	req := &v1.SubscribeRequest{
		StartDelay: uint32(*startDelay),
		Frequency:  uint32(*frequency),
		StatTypes:  stats,
	}

	if *verbose {
		log.Printf("Connecting to %s", *serverAddr)
		log.Printf("Request: delay=%ds, frequency=%ds, stat_types=%v", req.StartDelay, req.Frequency, stats)
	}

	// Start subscription
	stream, err := client.Subscribe(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}

	log.Printf("Subscribed to PulseCat server. Receiving stats every %d seconds...", *frequency)
	if *duration > 0 {
		log.Printf("Will run for %d seconds", *duration)
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Set up duration timer if specified
	var durationTimer *time.Timer
	if *duration > 0 {
		durationTimer = time.NewTimer(time.Duration(*duration) * time.Second)
		defer durationTimer.Stop()
	}

	statsCount := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sigCh:
			log.Printf("Interrupt received, shutting down...")
			return nil
		case <-durationTimer.C:
			log.Printf("Duration reached, shutting down...")
			return nil
		default:
			// Receive next stats
			stats, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Server closed stream")
				return nil
			}
			if err != nil {
				return fmt.Errorf("error receiving stats: %v", err)
			}

			statsCount++
			printStats(stats)

			if *verbose {
				log.Printf("Received stat batch #%d", statsCount)
			}
		}
	}
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

	log.Printf("PulseKitten client finished")
}
