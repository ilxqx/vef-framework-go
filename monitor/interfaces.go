package monitor

import "context"

// Service defines the interface for system monitoring operations.
// It provides methods to retrieve various system metrics and information
// including cpu, memory, disk, network, host, process, load, and build info.
type Service interface {
	// Overview returns a comprehensive system overview including all metrics.
	// This method aggregates cpu, memory, disk, network, host, process, load, and build info.
	Overview(ctx context.Context) (*SystemOverview, error)
	// Cpu returns detailed cpu information including usage percentages.
	// Note: This method may take ~1 second due to cpu usage sampling.
	Cpu(ctx context.Context) (*CpuInfo, error)
	// Memory returns memory usage information including virtual and swap memory.
	Memory(ctx context.Context) (*MemoryInfo, error)
	// Disk returns disk usage and partition information.
	Disk(ctx context.Context) (*DiskInfo, error)
	// Network returns network interface and I/O statistics.
	Network(ctx context.Context) (*NetworkInfo, error)
	// Host returns static host information such as OS, platform, and kernel version.
	// This information is cached for 5 minutes to avoid unnecessary system calls.
	Host(ctx context.Context) (*HostInfo, error)
	// Process returns information about the current process.
	// Note: This method may take ~1 second due to cpu usage sampling.
	Process(ctx context.Context) (*ProcessInfo, error)
	// Load returns system load averages.
	Load(ctx context.Context) (*LoadInfo, error)
	// BuildInfo returns application build information if available.
	// Returns nil if no build info was provided during service creation.
	// The FrameworkVersion field is automatically populated by the framework.
	BuildInfo() *BuildInfo
}
