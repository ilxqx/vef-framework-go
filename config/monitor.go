package config

import "time"

// MonitorConfig defines the configuration for the monitoring service.
type MonitorConfig struct {
	// SampleInterval is the interval between CPU and process sampling.
	// Default: 10 seconds
	SampleInterval time.Duration `config:"sample_interval"`
	// SampleDuration is the sampling window duration for CPU and process metrics.
	// Default: 2 seconds
	SampleDuration time.Duration `config:"sample_duration"`
}
