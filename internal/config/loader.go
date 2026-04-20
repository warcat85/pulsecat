package config

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// parse command-line flags and YAML configuration file.
// merge them appropriately (flags override YAML).
// Return the parsed configuration and any error.
func LoadConfig() (*Config, error) {
	var config Config
	var configFile string

	// Command-line flags (overrides config file)
	flag.IntVar(
		&config.Port, "port", DefaultPort, "Port for gRPC server (overrides config file)")
	flag.StringVar(
		&config.LogLevel, "log-level", DefaultLogLevel, "Log level (debug, info, warn, error) (overrides config file)")
	flag.StringVar(
		&configFile, "config", "configs/config.yaml", "Path to YAML configuration file")

	// Add help flag
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help {
		PrintUsage()
		os.Exit(0)
	}

	var portSpecified, logLevelSpecified bool
	var yamlConfig *Config
	// Load configuration from YAML file if it exists
	if _, err := os.Stat(configFile); err == nil {
		yamlConfig, err = loadFromFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configFile, err)
		}
		flag.Visit(func(f *flag.Flag) {
			name := f.Name
			if name == "port" {
				log.Printf("Overriding port from config file with command-line value: %d", config.Port)
				portSpecified = true
			} else if name == "log-level" {
				log.Printf("Overriding log level from config file with command-line value: %s", config.LogLevel)
				logLevelSpecified = true
			}
		})
	}

	// Merge YAML config with command-line flags (flags override YAML)
	if yamlConfig != nil {
		if !portSpecified {
			config.Port = yamlConfig.Port
		}
		if !logLevelSpecified {
			config.LogLevel = yamlConfig.LogLevel
		}

		// using validated YAML values
		config.CollectionInterval = yamlConfig.CollectionInterval
		config.BufferDuration = yamlConfig.BufferDuration
		config.Monitors = yamlConfig.Monitors

		log.Printf("Loaded configuration from %s", configFile)
	} else {
		config.CollectionInterval = DefaultCollectionInterval
		config.BufferDuration = DefaultBufferDuration
		config.Monitors = allMonitors()
		log.Printf("Config file %s not found, using defaults and command-line flags", configFile)
	}

	// Validate configuration
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}
	if config.BufferDuration < config.CollectionInterval {
		return nil, fmt.Errorf("buffer duration must be at least collection interval")
	}

	enabledMonitors := config.Monitors.Count()
	log.Printf(
		"Configuration loaded: "+
			"port=%d, log-level=%s, collection-interval=%ds, buffer-duration=%ds, monitors=%d/7 enabled",
		config.Port, config.LogLevel, config.CollectionInterval, config.BufferDuration, enabledMonitors)

	return &config, nil
}

// prints the command-line usage information.
func PrintUsage() {
	programName := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", programName)
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

// load configuration from a YAML file.
func loadFromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config YAMLConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	collectionInterval := config.Advanced.CollectionInterval
	if collectionInterval <= 0 {
		log.Printf(
			"Invalid collection interval %d, using default %d",
			collectionInterval, DefaultCollectionInterval)
		collectionInterval = DefaultCollectionInterval
	}

	bufferDuration := config.Advanced.BufferDuration
	if bufferDuration <= 0 {
		log.Printf("Invalid buffer duration %d, using default %d", bufferDuration, DefaultBufferDuration)
		bufferDuration = DefaultBufferDuration
	}

	monitors := config.Monitors
	if monitors.IsEmpty() {
		monitors = allMonitors()
		log.Printf("No monitor configuration found, enabling all monitors by default")
	}

	// Convert YAML config to internal Config
	return &Config{
		Port:               config.Server.Port,
		LogLevel:           config.Logging.Level,
		Monitors:           monitors,
		CollectionInterval: collectionInterval,
		BufferDuration:     bufferDuration,
	}, nil
}
