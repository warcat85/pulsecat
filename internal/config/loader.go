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

// parse command-line flags and YAML configuration file,
// merge them appropriately (flags override YAML).
// Return the parsed configuration and any error.
func LoadConfig() (*Config, error) {
	var config Config
	var configFile string

	// Command-line flags (overrides config file)
	flag.IntVar(&config.Port, "port", DefaultPort, "Port for gRPC server (overrides config file)")
	flag.StringVar(&config.LogLevel, "log-level", DefaultLogLevel, "Log level (debug, info, warn, error) (overrides config file)")
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to YAML configuration file")

	// Add help flag
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help {
		PrintUsage()
		os.Exit(0)
	}

	// Track which flags were explicitly set
	specifiedFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		specifiedFlags[f.Name] = true
	})

	// Load configuration from YAML file if it exists
	var yamlConfig *Config
	if _, err := os.Stat(configFile); err == nil {
		var err error
		yamlConfig, err = loadFromFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %v", configFile, err)
		}
		// Merge YAML config with command-line flags (flags override YAML)
		if yamlConfig != nil {
			if !specifiedFlags["port"] {
				config.Port = yamlConfig.Port
			} else {
				log.Printf("Overriding port from config file with command-line value: %d", config.Port)
			}
			if !specifiedFlags["log-level"] {
				config.LogLevel = yamlConfig.LogLevel
			} else {
				log.Printf("Overriding log level from config file with command-line value: %s", config.LogLevel)
			}
			// use YAML value with validation
			if yamlConfig.CollectionInterval <= 0 {
				log.Printf("Invalid collection interval %d, using default %d", yamlConfig.CollectionInterval, DefaultCollectionInterval)
				config.CollectionInterval = DefaultCollectionInterval
			} else {
				config.CollectionInterval = yamlConfig.CollectionInterval
			}
			// use YAML value with validation
			if yamlConfig.BufferDuration <= 0 {
				log.Printf("Invalid buffer duration %d, using default %d", yamlConfig.BufferDuration, DefaultBufferDuration)
				config.BufferDuration = DefaultBufferDuration
			} else {
				config.BufferDuration = yamlConfig.BufferDuration
			}
			if yamlConfig.Monitors.IsEmpty() {
				config.Monitors = allMonitors()
				log.Printf("No monitor configuration found, enabling all monitors by default")
			} else {
				config.Monitors = yamlConfig.Monitors
			}
			log.Printf("Loaded configuration from %s", configFile)
		}
	} else {
		log.Printf("Config file %s not found, using defaults and command-line flags", configFile)
		config.CollectionInterval = DefaultCollectionInterval
		config.BufferDuration = DefaultBufferDuration
		config.Monitors = allMonitors()
	}

	// Validate configuration
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}
	if config.BufferDuration < config.CollectionInterval {
		return nil, fmt.Errorf("buffer duration must be at least collection interval")
	}

	enabledMonitors := config.Monitors.Count()
	log.Printf("Configuration loaded: port=%d, log-level=%s, collection-interval=%ds, buffer-duration=%ds, monitors=%d/7 enabled",
		config.Port, config.LogLevel, config.CollectionInterval, config.BufferDuration, enabledMonitors)

	return &config, nil
}

// prints the command-line usage information
func PrintUsage() {
	programName := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", programName)
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

// load configuration from a YAML file
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

	var yamlConfig YAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Convert YAML config to internal Config
	config := &Config{
		Port:               yamlConfig.Server.Port,
		LogLevel:           yamlConfig.Logging.Level,
		Monitors:           yamlConfig.Monitors,
		CollectionInterval: yamlConfig.Advanced.CollectionInterval,
		BufferDuration:     yamlConfig.Advanced.BufferDuration,
	}

	return config, nil
}
