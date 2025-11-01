package monitor

// SystemOverview provides a comprehensive snapshot of all system metrics.
type SystemOverview struct {
	// Host information summary
	Host *HostSummary `json:"host"`
	// Cpu metrics summary
	Cpu *CpuSummary `json:"cpu"`
	// Memory metrics summary
	Memory *MemorySummary `json:"memory"`
	// Disk metrics summary
	Disk *DiskSummary `json:"disk"`
	// Network metrics summary
	Network *NetworkSummary `json:"network"`
	// Current process metrics summary
	Process *ProcessSummary `json:"process"`
	// System load averages
	Load *LoadInfo `json:"load"`
	// Application build information (may be nil if not provided)
	Build *BuildInfo `json:"build"`
}

// HostSummary provides a summary of host information.
type HostSummary struct {
	// Hostname of the machine
	Hostname string `json:"hostname"`
	// Operating system name (e.g., "darwin", "linux", "windows")
	Os string `json:"os"`
	// Platform name (e.g., "ubuntu", "centos", "darwin")
	Platform string `json:"platform"`
	// System uptime in seconds
	UpTime uint64 `json:"upTime"`
}

// HostInfo contains detailed static information about the host system.
type HostInfo struct {
	// Hostname of the machine
	Hostname string `json:"hostname"`
	// System uptime in seconds
	UpTime uint64 `json:"upTime"`
	// System boot time as Unix timestamp in seconds
	BootTime uint64 `json:"bootTime"`
	// Number of running processes
	Processes uint64 `json:"processes"`
	// Operating system name (e.g., "darwin", "linux", "windows")
	Os string `json:"os"`
	// Platform name (e.g., "ubuntu", "centos", "darwin")
	Platform string `json:"platform"`
	// Platform family (e.g., "debian", "rhel", "standalone")
	PlatformFamily string `json:"platformFamily"`
	// Platform version string
	PlatformVersion string `json:"platformVersion"`
	// Kernel version string
	KernelVersion string `json:"kernelVersion"`
	// Kernel architecture (e.g., "x86_64", "arm64")
	KernelArch string `json:"kernelArch"`
	// Virtualization system (e.g., "kvm", "xen", "vmware")
	VirtualizationSystem string `json:"virtualizationSystem"`
	// Virtualization role (e.g., "host", "guest")
	VirtualizationRole string `json:"virtualizationRole"`
	// Unique host identifier
	HostId string `json:"hostId"`
}

// CpuSummary provides a summary of cpu metrics for the overview.
type CpuSummary struct {
	// Number of physical cpu cores
	PhysicalCores int `json:"physicalCores"`
	// Number of logical cpu cores (includes hyperthreading)
	LogicalCores int `json:"logicalCores"`
	// Overall cpu usage percentage (0-100 * logical cores)
	UsagePercent float64 `json:"usagePercent"`
}

// CpuInfo contains detailed cpu information including per-core usage.
type CpuInfo struct {
	// Number of physical cpu cores
	PhysicalCores int `json:"physicalCores"`
	// Number of logical cpu cores (includes hyperthreading)
	LogicalCores int `json:"logicalCores"`
	// Cpu model name (e.g., "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz")
	ModelName string `json:"modelName"`
	// Cpu frequency in MHz
	Mhz float64 `json:"mhz"`
	// Cpu cache size in KB
	CacheSize int32 `json:"cacheSize"`
	// Per-core cpu usage percentage (0-100 for each core)
	UsagePercent []float64 `json:"usagePercent"`
	// Total cpu usage percentage across all cores (0-100)
	TotalPercent float64 `json:"totalPercent"`
	// Cpu vendor ID (e.g., "GenuineIntel", "AuthenticAMD")
	VendorId string `json:"vendorId"`
	// Cpu family identifier
	Family string `json:"family"`
	// Cpu model identifier
	Model string `json:"model"`
	// Cpu stepping number
	Stepping int32 `json:"stepping"`
	// Cpu microcode version
	Microcode string `json:"microcode"`
}

// MemorySummary provides a summary of memory metrics for the overview.
type MemorySummary struct {
	// Total memory in bytes
	Total uint64 `json:"total"`
	// Used memory in bytes
	Used uint64 `json:"used"`
	// Memory usage percentage (0-100)
	UsedPercent float64 `json:"usedPercent"`
}

// MemoryInfo contains detailed memory information.
type MemoryInfo struct {
	// Virtual (physical) memory statistics
	Virtual *VirtualMemory `json:"virtual"`
	// Swap memory statistics
	Swap *SwapMemory `json:"swap"`
}

// VirtualMemory represents virtual (physical) memory statistics.
type VirtualMemory struct {
	// Total physical memory in bytes
	Total uint64 `json:"total"`
	// Available memory in bytes (free + buffers + cache)
	Available uint64 `json:"available"`
	// Used memory in bytes
	Used uint64 `json:"used"`
	// Memory usage percentage (0-100)
	UsedPercent float64 `json:"usedPercent"`
	// Free memory in bytes
	Free uint64 `json:"free"`
	// Active memory in bytes (recently used)
	Active uint64 `json:"active"`
	// Inactive memory in bytes (not recently used)
	Inactive uint64 `json:"inactive"`
	// Wired memory in bytes (cannot be paged out, macOS/BSD)
	Wired uint64 `json:"wired"`
	// Laundry memory in bytes (queued for cleaning, BSD)
	Laundry uint64 `json:"laundry"`
	// Buffer cache memory in bytes
	Buffers uint64 `json:"buffers"`
	// Page cache memory in bytes
	Cached uint64 `json:"cached"`
	// Memory waiting to be written back to disk in bytes
	WriteBack uint64 `json:"writeBack"`
	// Dirty memory in bytes (modified but not written to disk)
	Dirty uint64 `json:"dirty"`
	// Temporary writeback memory in bytes
	WriteBackTmp uint64 `json:"writeBackTmp"`
	// Shared memory in bytes (shared between processes)
	Shared uint64 `json:"shared"`
	// Total slab memory in bytes (kernel data structures)
	Slab uint64 `json:"slab"`
	// Reclaimable slab memory in bytes (can be freed)
	SlabReclaimable uint64 `json:"slabReclaimable"`
	// Unreclaimable slab memory in bytes (cannot be freed)
	SlabUnreclaimable uint64 `json:"slabUnreclaimable"`
	// Memory used for page tables in bytes
	PageTables uint64 `json:"pageTables"`
	// Swap cached memory in bytes
	SwapCached uint64 `json:"swapCached"`
	// Commit limit in bytes (max memory that can be allocated)
	CommitLimit uint64 `json:"commitLimit"`
	// Committed memory in bytes (allocated virtual memory)
	CommittedAs uint64 `json:"committedAs"`
	// Total high memory in bytes (Linux 32-bit)
	HighTotal uint64 `json:"highTotal"`
	// Free high memory in bytes (Linux 32-bit)
	HighFree uint64 `json:"highFree"`
	// Total low memory in bytes (Linux 32-bit)
	LowTotal uint64 `json:"lowTotal"`
	// Free low memory in bytes (Linux 32-bit)
	LowFree uint64 `json:"lowFree"`
	// Total swap space in bytes
	SwapTotal uint64 `json:"swapTotal"`
	// Free swap space in bytes
	SwapFree uint64 `json:"swapFree"`
	// Mapped memory in bytes (memory-mapped files)
	Mapped uint64 `json:"mapped"`
	// Total virtual memory allocation space in bytes
	VmAllocTotal uint64 `json:"vmAllocTotal"`
	// Used virtual memory allocation space in bytes
	VmAllocUsed uint64 `json:"vmAllocUsed"`
	// Largest contiguous virtual memory allocation chunk in bytes
	VmAllocChunk uint64 `json:"vmAllocChunk"`
	// Total huge pages count
	HugePagesTotal uint64 `json:"hugePagesTotal"`
	// Free huge pages count
	HugePagesFree uint64 `json:"hugePagesFree"`
	// Reserved huge pages count
	HugePagesReserved uint64 `json:"hugePagesReserved"`
	// Surplus huge pages count
	HugePagesSurplus uint64 `json:"hugePagesSurplus"`
	// Size of each huge page in bytes
	HugePageSize uint64 `json:"hugePageSize"`
	// Anonymous huge pages in bytes (transparent huge pages)
	AnonHugePages uint64 `json:"anonHugePages"`
}

// SwapMemory represents swap memory statistics.
type SwapMemory struct {
	// Total swap space in bytes
	Total uint64 `json:"total"`
	// Used swap space in bytes
	Used uint64 `json:"used"`
	// Free swap space in bytes
	Free uint64 `json:"free"`
	// Swap usage percentage (0-100)
	UsedPercent float64 `json:"usedPercent"`
	// Number of bytes swapped in from disk
	SwapIn uint64 `json:"swapIn"`
	// Number of bytes swapped out to disk
	SwapOut uint64 `json:"swapOut"`
	// Number of pages swapped in from disk
	PageIn uint64 `json:"pageIn"`
	// Number of pages swapped out to disk
	PageOut uint64 `json:"pageOut"`
	// Total number of page faults
	PageFault uint64 `json:"pageFault"`
	// Number of major page faults (required disk I/O)
	PageMajorFault uint64 `json:"pageMajorFault"`
}

// DiskSummary provides a summary of disk metrics for the overview.
type DiskSummary struct {
	// Total disk space in bytes (sum of all partitions)
	Total uint64 `json:"total"`
	// Used disk space in bytes (sum of all partitions)
	Used uint64 `json:"used"`
	// Disk usage percentage (0-100)
	UsedPercent float64 `json:"usedPercent"`
	// Number of disk partitions
	Partitions int `json:"partitions"`
}

// DiskInfo contains detailed disk information including partitions and I/O counters.
type DiskInfo struct {
	// List of disk partitions
	Partitions []*PartitionInfo `json:"partitions"`
	// Disk I/O statistics per device
	IoCounters map[string]*IoCounter `json:"ioCounters"`
}

// PartitionInfo represents a disk partition.
type PartitionInfo struct {
	// Device name (e.g., "/dev/sda1")
	Device string `json:"device"`
	// Mount point path (e.g., "/", "/home")
	MountPoint string `json:"mountPoint"`
	// File system type (e.g., "ext4", "ntfs", "apfs")
	FsType string `json:"fsType"`
	// Mount options (e.g., ["rw", "relatime"])
	Options []string `json:"options"`
	// Total partition size in bytes
	Total uint64 `json:"total"`
	// Free space in bytes
	Free uint64 `json:"free"`
	// Used space in bytes
	Used uint64 `json:"used"`
	// Disk usage percentage (0-100)
	UsedPercent float64 `json:"usedPercent"`
	// Total number of index nodes (inodes)
	INodesTotal uint64 `json:"iNodesTotal"`
	// Number of used index nodes
	INodesUsed uint64 `json:"iNodesUsed"`
	// Number of free index nodes
	INodesFree uint64 `json:"iNodesFree"`
	// Index nodes usage percentage (0-100)
	INodesUsedPercent float64 `json:"iNodesUsedPercent"`
}

// IoCounter represents disk I/O statistics.
type IoCounter struct {
	// Number of read operations
	ReadCount uint64 `json:"readCount"`
	// Number of merged read operations
	MergedReadCount uint64 `json:"mergedReadCount"`
	// Number of write operations
	WriteCount uint64 `json:"writeCount"`
	// Number of merged write operations
	MergedWriteCount uint64 `json:"mergedWriteCount"`
	// Total bytes read
	ReadBytes uint64 `json:"readBytes"`
	// Total bytes written
	WriteBytes uint64 `json:"writeBytes"`
	// Total time spent reading in milliseconds
	ReadTime uint64 `json:"readTime"`
	// Total time spent writing in milliseconds
	WriteTime uint64 `json:"writeTime"`
	// Number of I/O operations currently in progress
	IopsInProgress uint64 `json:"iopsInProgress"`
	// Total time spent doing I/Os in milliseconds
	IoTime uint64 `json:"ioTime"`
	// Weighted time spent doing I/Os in milliseconds
	WeightedIo uint64 `json:"weightedIo"`
	// Device name
	Name string `json:"name"`
	// Device serial number
	SerialNumber string `json:"serialNumber"`
	// Device label
	Label string `json:"label"`
}

// NetworkSummary provides a summary of network metrics for the overview.
type NetworkSummary struct {
	// Number of network interfaces
	Interfaces int `json:"interfaces"`
	// Total bytes sent across all interfaces
	BytesSent uint64 `json:"bytesSent"`
	// Total bytes received across all interfaces
	BytesRecv uint64 `json:"bytesRecv"`
	// Total packets sent across all interfaces
	PacketsSent uint64 `json:"packetsSent"`
	// Total packets received across all interfaces
	PacketsRecv uint64 `json:"packetsRecv"`
}

// NetworkInfo contains detailed network interface and I/O information.
type NetworkInfo struct {
	// List of network interfaces
	Interfaces []*InterfaceInfo `json:"interfaces"`
	// Network I/O statistics per interface
	IoCounters map[string]*NetIoCounter `json:"ioCounters"`
}

// InterfaceInfo represents a network interface.
type InterfaceInfo struct {
	// Interface index number
	Index int `json:"index"`
	// Maximum transmission unit in bytes
	Mtu int `json:"mtu"`
	// Interface name (e.g., "eth0", "wlan0", "en0")
	Name string `json:"name"`
	// Hardware MAC address
	HardwareAddr string `json:"hardwareAddr"`
	// Interface flags (e.g., ["up", "broadcast", "multicast"])
	Flags []string `json:"flags"`
	// IP addresses assigned to this interface
	Addrs []string `json:"addrs"`
}

// NetIoCounter represents network I/O statistics.
type NetIoCounter struct {
	// Interface name
	Name string `json:"name"`
	// Total bytes sent
	BytesSent uint64 `json:"bytesSent"`
	// Total bytes received
	BytesRecv uint64 `json:"bytesRecv"`
	// Total packets sent
	PacketsSent uint64 `json:"packetsSent"`
	// Total packets received
	PacketsRecv uint64 `json:"packetsRecv"`
	// Total incoming errors
	ErrorsIn uint64 `json:"errorsIn"`
	// Total outgoing errors
	ErrorsOut uint64 `json:"errorsOut"`
	// Total incoming packets dropped
	DroppedIn uint64 `json:"droppedIn"`
	// Total outgoing packets dropped
	DroppedOut uint64 `json:"droppedOut"`
	// Total incoming FIFO buffer errors
	FifoIn uint64 `json:"fifoIn"`
	// Total outgoing FIFO buffer errors
	FifoOut uint64 `json:"fifoOut"`
}

// ProcessSummary provides a summary of process metrics for the overview.
type ProcessSummary struct {
	// Process ID
	Pid int32 `json:"pid"`
	// Process name
	Name string `json:"name"`
	// Cpu usage percentage (0-100)
	CpuPercent float64 `json:"cpuPercent"`
	// Memory usage percentage (0-100)
	MemoryPercent float32 `json:"memoryPercent"`
}

// ProcessInfo contains detailed information about the current process.
type ProcessInfo struct {
	// Process ID
	Pid int32 `json:"pid"`
	// Parent process ID
	ParentPid int32 `json:"parentPid"`
	// Process name
	Name string `json:"name"`
	// Executable path
	Exe string `json:"exe"`
	// Command line with arguments
	Cmdline string `json:"cmdline"`
	// Current working directory
	Cwd string `json:"cwd"`
	// Process status (e.g., "running", "sleeping", "stopped")
	Status string `json:"status"`
	// Username of process owner
	Username string `json:"username"`
	// Process creation time as Unix timestamp in milliseconds
	CreateTime int64 `json:"createTime"`
	// Number of threads
	NumThreads int32 `json:"numThreads"`
	// Number of file descriptors
	NumFds int32 `json:"numFds"`
	// Cpu usage percentage (0-100)
	CpuPercent float64 `json:"cpuPercent"`
	// Memory usage percentage (0-100)
	MemoryPercent float32 `json:"memoryPercent"`
	// Resident set size (physical memory) in bytes
	MemoryRss uint64 `json:"memoryRss"`
	// Virtual memory size in bytes
	MemoryVms uint64 `json:"memoryVms"`
	// Swap memory usage in bytes
	MemorySwap uint64 `json:"memorySwap"`
}

// LoadInfo represents system load averages.
type LoadInfo struct {
	// 1-minute load average (number of processes in run queue)
	Load1 float64 `json:"load1"`
	// 5-minute load average (number of processes in run queue)
	Load5 float64 `json:"load5"`
	// 15-minute load average (number of processes in run queue)
	Load15 float64 `json:"load15"`
}

// BuildInfo contains application build metadata.
// The FrameworkVersion field is automatically populated by the framework
// and should not be set by users.
type BuildInfo struct {
	// VEF Framework version (automatically populated, e.g., "v0.7.2")
	// This field is set by the framework and will override any user-provided value
	VEFVersion string `json:"vefVersion"`
	// Application version (e.g., "v1.0.0")
	AppVersion string `json:"appVersion"`
	// Build timestamp (e.g., "2024-01-01T00:00:00Z")
	BuildTime string `json:"buildTime"`
	// Git commit hash (e.g., "abc123def456")
	GitCommit string `json:"gitCommit"`
}
