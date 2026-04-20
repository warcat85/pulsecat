package config

// default values
const (
	DefaultPort               = 25225 // port for gRPC server (BLACK in numbers)
	DefaultLogLevel           = "info"
	DefaultCollectionInterval = 1   // seconds
	DefaultBufferDuration     = 300 // seconds (5 minutes)
)

// application configuration
type Config struct {
	Port               int // Port for gRPC server
	LogLevel           string
	CollectionInterval int // Interval between metric collections (seconds)
	BufferDuration     int // Maximum window size for sliding window (seconds)

	// Monitor enable/disable settings
	Monitors MonitorsConfig
}

// enable/disable settings for each monitor type
type MonitorsConfig struct {
	LoadAverage         bool `yaml:"load_average"`
	CPUUsage            bool `yaml:"cpu_usage"`
	DiskUsage           bool `yaml:"disk_usage"`
	NetworkStats        bool `yaml:"network_stats"`
	TopTalkers          bool `yaml:"top_talkers"`
	ListeningSockets    bool `yaml:"listening_sockets"`
	TCPConnectionStates bool `yaml:"tcp_connection_states"`
}

// number of enabled monitors
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

// returns true if no monitors are enabled
func (m MonitorsConfig) IsEmpty() bool {
	return m.Count() == 0
}

// new MonitorsConfig with all monitors enabled
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

// configuration in YAML file
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
	CollectionInterval int `yaml:"collection_interval,omitempty"`
	BufferDuration     int `yaml:"buffer_duration,omitempty"`
	// TODO configuration for future improvement
	MaxTopTalkers      int      `yaml:"max_top_talkers,omitempty"`
	NetworkInterface   string   `yaml:"network_interface,omitempty"`
	ExcludeFilesystems []string `yaml:"exclude_filesystems,omitempty"`
}
