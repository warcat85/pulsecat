package config

// default values.
const (
	DefaultPort               = 25225 // port for gRPC server (BLACK in numbers)
	DefaultLogLevel           = "info"
	DefaultCollectionInterval = 1   // seconds
	DefaultBufferDuration     = 300 // seconds (5 minutes)
)

// application configuration.
type Config struct {
	Port               int // Port for gRPC server
	LogLevel           string
	CollectionInterval int // Interval between metric collections (seconds)
	BufferDuration     int // Maximum window size for sliding window (seconds)

	// Monitor enable/disable settings
	Monitors MonitorsConfig
}

// enable/disable each monitor type.
type MonitorsConfig struct {
	LoadAverage         bool `yaml:"loadAverage"`
	CPUUsage            bool `yaml:"cpuUsage"`
	DiskUsage           bool `yaml:"diskUsage"`
	NetworkStats        bool `yaml:"networkStats"`
	TopTalkers          bool `yaml:"topTalkers"`
	ListeningSockets    bool `yaml:"listeningSockets"`
	TCPConnectionStates bool `yaml:"tcpConnectionStates"`
}

// number of enabled monitors.
func (m MonitorsConfig) Count() int {
	count := 0
	if m.LoadAverage {
		count++
	}
	if m.CPUUsage {
		count++
	}
	if m.DiskUsage {
		count++
	}
	if m.NetworkStats {
		count++
	}
	if m.TopTalkers {
		count++
	}
	if m.ListeningSockets {
		count++
	}
	if m.TCPConnectionStates {
		count++
	}
	return count
}

// returns true if no monitors are enabled.
func (m MonitorsConfig) IsEmpty() bool {
	return m.Count() == 0
}

// new MonitorsConfig with all monitors enabled.
func allMonitors() MonitorsConfig {
	return MonitorsConfig{
		LoadAverage:         true,
		CPUUsage:            true,
		DiskUsage:           true,
		NetworkStats:        true,
		TopTalkers:          true,
		ListeningSockets:    true,
		TCPConnectionStates: true,
	}
}

// configuration in YAML file.
type YAMLConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Monitors MonitorsConfig `yaml:"monitors"`
	Advanced AdvancedConfig `yaml:"advanced,omitempty"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file,omitempty"`
}

type AdvancedConfig struct {
	CollectionInterval int `yaml:"collectionInterval,omitempty"`
	BufferDuration     int `yaml:"bufferDuration,omitempty"`
	// TODO configuration for future improvement
	MaxTopTalkers      int      `yaml:"maxTopTalkers,omitempty"`
	NetworkInterface   string   `yaml:"networkInterface,omitempty"`
	ExcludeFilesystems []string `yaml:"excludeFilesystems,omitempty"`
}
