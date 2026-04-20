package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"pulsecat/internal/app"
	"pulsecat/internal/config"
	"pulsecat/internal/version"
	"syscall"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Starting PulseCat daemon v%s (build: %s, commit: %s)",
		version.Version, version.BuildTime, version.CommitHash)

	// Parse configuration (includes command-line flags and YAML file)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Create daemon
	daemon := app.NewDaemon(cfg)

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Start daemon in a goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Println("Daemon starting...")
		if err := daemon.Run(); err != nil {
			errCh <- fmt.Errorf("daemon failed: %w", err)
		} else {
			errCh <- nil
		}
	}()

	// Wait for either a signal or daemon error
	select {
	case sig := <-sigCh:
		log.Printf("Received signal: %v", sig)
		log.Println("Initiating graceful shutdown...")
		daemon.Stop()
		log.Println("Shutdown complete")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Fatal error: %v", err)
		}
		log.Println("Daemon stopped normally")
	}

	log.Println("System monitor daemon exited")
}
