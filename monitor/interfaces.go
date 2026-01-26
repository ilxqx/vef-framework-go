package monitor

import "context"

// Service defines the interface for system monitoring operations.
type Service interface {
	// Overview returns a comprehensive system overview including all metrics.
	Overview(ctx context.Context) (*SystemOverview, error)
	// CPU returns detailed CPU information including usage percentages.
	CPU(ctx context.Context) (*CPUInfo, error)
	// Memory returns memory usage information including virtual and swap memory.
	Memory(ctx context.Context) (*MemoryInfo, error)
	// Disk returns disk usage and partition information.
	Disk(ctx context.Context) (*DiskInfo, error)
	// Network returns network interface and I/O statistics.
	Network(ctx context.Context) (*NetworkInfo, error)
	// Host returns static host information such as OS, platform, and kernel version.
	Host(ctx context.Context) (*HostInfo, error)
	// Process returns information about the current process.
	Process(ctx context.Context) (*ProcessInfo, error)
	// Load returns system load averages.
	Load(ctx context.Context) (*LoadInfo, error)
	// BuildInfo returns application build information if available.
	BuildInfo() *BuildInfo
}
