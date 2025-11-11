package monitor

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/monitor"
)

type DefaultService struct {
	buildInfo *monitor.BuildInfo
	config    *config.MonitorConfig

	// Atomic caches for cpu and process metrics
	cpuCache     atomic.Value // stores *monitor.CpuInfo
	processCache atomic.Value // stores *monitor.ProcessInfo

	// Background sampler control
	samplerCancel context.CancelFunc
	samplerDone   chan struct{}
}

// NewService creates a new monitor.Service implementation.
func NewService(
	cfg *config.MonitorConfig,
	buildInfo *monitor.BuildInfo,
) monitor.Service {
	return &DefaultService{
		buildInfo: buildInfo,
		config:    cfg,
	}
}

// Overview returns a comprehensive system overview by fetching all metrics.
func (s *DefaultService) Overview(ctx context.Context) (*monitor.SystemOverview, error) {
	var overview monitor.SystemOverview

	// Host info
	if hostInfo, err := s.Host(ctx); err == nil {
		overview.Host = &monitor.HostSummary{
			Hostname: hostInfo.Hostname,
			Os:       hostInfo.Os,
			Platform: hostInfo.Platform,
			UpTime:   hostInfo.UpTime,
		}
	}

	// Cpu info
	if cpuInfo, err := s.Cpu(ctx); err == nil {
		overview.Cpu = &monitor.CpuSummary{
			PhysicalCores: cpuInfo.PhysicalCores,
			LogicalCores:  cpuInfo.LogicalCores,
			UsagePercent:  cpuInfo.TotalPercent,
		}
	}

	// Memory info
	if memInfo, err := s.Memory(ctx); err == nil && memInfo.Virtual != nil {
		overview.Memory = &monitor.MemorySummary{
			Total:       memInfo.Virtual.Total,
			Used:        memInfo.Virtual.Used,
			UsedPercent: memInfo.Virtual.UsedPercent,
		}
	}

	// Disk info
	if diskInfo, err := s.Disk(ctx); err == nil {
		var (
			total, used uint64
			usedPercent float64
		)

		for _, part := range diskInfo.Partitions {
			total += part.Total
			used += part.Used
		}

		if total > 0 {
			usedPercent = float64(used) / float64(total) * 100
		}

		overview.Disk = &monitor.DiskSummary{
			Total:       total,
			Used:        used,
			UsedPercent: usedPercent,
			Partitions:  len(diskInfo.Partitions),
		}
	}

	// Network info
	if netInfo, err := s.Network(ctx); err == nil {
		var bytesSent, bytesRecv, packetsSent, packetsRecv uint64
		for _, counter := range netInfo.IoCounters {
			bytesSent += counter.BytesSent
			bytesRecv += counter.BytesRecv
			packetsSent += counter.PacketsSent
			packetsRecv += counter.PacketsRecv
		}

		overview.Network = &monitor.NetworkSummary{
			Interfaces:  len(netInfo.Interfaces),
			BytesSent:   bytesSent,
			BytesRecv:   bytesRecv,
			PacketsSent: packetsSent,
			PacketsRecv: packetsRecv,
		}
	}

	// Process info
	if procInfo, err := s.Process(ctx); err == nil {
		overview.Process = &monitor.ProcessSummary{
			Pid:           procInfo.Pid,
			Name:          procInfo.Name,
			CpuPercent:    procInfo.CpuPercent,
			MemoryPercent: procInfo.MemoryPercent,
		}
	}

	// Load info
	if loadInfo, err := s.Load(ctx); err == nil {
		overview.Load = loadInfo
	}

	// Build info
	overview.Build = s.BuildInfo()

	return &overview, nil
}

// Cpu returns detailed cpu information including usage percentages.
// This method returns cached data from background sampling, ensuring fast response.
func (s *DefaultService) Cpu(_ context.Context) (*monitor.CpuInfo, error) {
	cached := s.cpuCache.Load()
	if cached == nil {
		return nil, ErrCpuInfoNotReady
	}

	return cached.(*monitor.CpuInfo), nil
}

// Memory returns memory usage information.
func (s *DefaultService) Memory(ctx context.Context) (*monitor.MemoryInfo, error) {
	// Get virtual memory
	vMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	virtualMemory := &monitor.VirtualMemory{
		Total:             vMem.Total,
		Available:         vMem.Available,
		Used:              vMem.Used,
		UsedPercent:       vMem.UsedPercent,
		Free:              vMem.Free,
		Active:            vMem.Active,
		Inactive:          vMem.Inactive,
		Wired:             vMem.Wired,
		Laundry:           vMem.Laundry,
		Buffers:           vMem.Buffers,
		Cached:            vMem.Cached,
		WriteBack:         vMem.WriteBack,
		Dirty:             vMem.Dirty,
		WriteBackTmp:      vMem.WriteBackTmp,
		Shared:            vMem.Shared,
		Slab:              vMem.Slab,
		SlabReclaimable:   vMem.Sreclaimable,
		SlabUnreclaimable: vMem.Sunreclaim,
		PageTables:        vMem.PageTables,
		SwapCached:        vMem.SwapCached,
		CommitLimit:       vMem.CommitLimit,
		CommittedAs:       vMem.CommittedAS,
		HighTotal:         vMem.HighTotal,
		HighFree:          vMem.HighFree,
		LowTotal:          vMem.LowTotal,
		LowFree:           vMem.LowFree,
		SwapTotal:         vMem.SwapTotal,
		SwapFree:          vMem.SwapFree,
		Mapped:            vMem.Mapped,
		VmAllocTotal:      vMem.VmallocTotal,
		VmAllocUsed:       vMem.VmallocUsed,
		VmAllocChunk:      vMem.VmallocChunk,
		HugePagesTotal:    vMem.HugePagesTotal,
		HugePagesFree:     vMem.HugePagesFree,
		HugePagesReserved: vMem.HugePagesRsvd,
		HugePagesSurplus:  vMem.HugePagesSurp,
		HugePageSize:      vMem.HugePageSize,
		AnonHugePages:     vMem.AnonHugePages,
	}

	// Get swap memory
	swapMem, err := mem.SwapMemoryWithContext(ctx)

	var swapMemory *monitor.SwapMemory
	if err == nil {
		swapMemory = &monitor.SwapMemory{
			Total:          swapMem.Total,
			Used:           swapMem.Used,
			Free:           swapMem.Free,
			UsedPercent:    swapMem.UsedPercent,
			SwapIn:         swapMem.Sin,
			SwapOut:        swapMem.Sout,
			PageIn:         swapMem.PgIn,
			PageOut:        swapMem.PgOut,
			PageFault:      swapMem.PgFault,
			PageMajorFault: swapMem.PgMajFault,
		}
	}

	return &monitor.MemoryInfo{
		Virtual: virtualMemory,
		Swap:    swapMemory,
	}, nil
}

// Disk returns disk usage and partition information.
func (s *DefaultService) Disk(ctx context.Context) (*monitor.DiskInfo, error) {
	// Get partitions
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	var partitionInfos []*monitor.PartitionInfo
	for _, part := range partitions {
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err != nil {
			continue
		}

		partitionInfos = append(partitionInfos, &monitor.PartitionInfo{
			Device:            part.Device,
			MountPoint:        part.Mountpoint,
			FsType:            part.Fstype,
			Options:           part.Opts,
			Total:             usage.Total,
			Free:              usage.Free,
			Used:              usage.Used,
			UsedPercent:       usage.UsedPercent,
			INodesTotal:       usage.InodesTotal,
			INodesUsed:        usage.InodesUsed,
			INodesFree:        usage.InodesFree,
			INodesUsedPercent: usage.InodesUsedPercent,
		})
	}

	// Get disk I/O counters (may not be supported on all platforms)
	ioCountersMap, err := disk.IOCountersWithContext(ctx)

	var ioCounters map[string]*monitor.IoCounter
	if err == nil {
		ioCounters = make(map[string]*monitor.IoCounter, len(ioCountersMap))
		for name, counter := range ioCountersMap {
			ioCounters[name] = &monitor.IoCounter{
				ReadCount:        counter.ReadCount,
				MergedReadCount:  counter.MergedReadCount,
				WriteCount:       counter.WriteCount,
				MergedWriteCount: counter.MergedWriteCount,
				ReadBytes:        counter.ReadBytes,
				WriteBytes:       counter.WriteBytes,
				ReadTime:         counter.ReadTime,
				WriteTime:        counter.WriteTime,
				IopsInProgress:   counter.IopsInProgress,
				IoTime:           counter.IoTime,
				WeightedIo:       counter.WeightedIO,
				Name:             counter.Name,
				SerialNumber:     counter.SerialNumber,
				Label:            counter.Label,
			}
		}
	}

	return &monitor.DiskInfo{
		Partitions: partitionInfos,
		IoCounters: ioCounters,
	}, nil
}

// Network returns network interface and I/O statistics.
func (s *DefaultService) Network(ctx context.Context) (*monitor.NetworkInfo, error) {
	// Get network interfaces
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var interfaceInfos []*monitor.InterfaceInfo
	for _, iface := range interfaces {
		var addrs []string
		for _, addr := range iface.Addrs {
			addrs = append(addrs, addr.Addr)
		}

		interfaceInfos = append(interfaceInfos, &monitor.InterfaceInfo{
			Index:        iface.Index,
			Mtu:          iface.MTU,
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr,
			Flags:        iface.Flags,
			Addrs:        addrs,
		})
	}

	// Get network I/O counters
	ioCountersMap, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	ioCounters := make(map[string]*monitor.NetIoCounter, len(ioCountersMap))
	for _, counter := range ioCountersMap {
		ioCounters[counter.Name] = &monitor.NetIoCounter{
			Name:        counter.Name,
			BytesSent:   counter.BytesSent,
			BytesRecv:   counter.BytesRecv,
			PacketsSent: counter.PacketsSent,
			PacketsRecv: counter.PacketsRecv,
			ErrorsIn:    counter.Errin,
			ErrorsOut:   counter.Errout,
			DroppedIn:   counter.Dropin,
			DroppedOut:  counter.Dropout,
			FifoIn:      counter.Fifoin,
			FifoOut:     counter.Fifoout,
		}
	}

	return &monitor.NetworkInfo{
		Interfaces: interfaceInfos,
		IoCounters: ioCounters,
	}, nil
}

// Host returns host information.
func (s *DefaultService) Host(ctx context.Context) (*monitor.HostInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return &monitor.HostInfo{
		Hostname:             info.Hostname,
		UpTime:               info.Uptime,
		BootTime:             info.BootTime,
		Processes:            info.Procs,
		Os:                   info.OS,
		Platform:             info.Platform,
		PlatformFamily:       info.PlatformFamily,
		PlatformVersion:      info.PlatformVersion,
		KernelVersion:        info.KernelVersion,
		KernelArch:           info.KernelArch,
		VirtualizationSystem: info.VirtualizationSystem,
		VirtualizationRole:   info.VirtualizationRole,
		HostId:               info.HostID,
	}, nil
}

// Process returns information about the current process.
// This method returns cached data from background sampling, ensuring fast response.
func (s *DefaultService) Process(_ context.Context) (*monitor.ProcessInfo, error) {
	cached := s.processCache.Load()
	if cached == nil {
		return nil, ErrProcessInfoNotReady
	}

	return cached.(*monitor.ProcessInfo), nil
}

// Load returns system load averages.
func (s *DefaultService) Load(ctx context.Context) (*monitor.LoadInfo, error) {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return &monitor.LoadInfo{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}

// BuildInfo returns application build information if available.
// Returns default "unknown" values if no build info was provided during service creation.
func (s *DefaultService) BuildInfo() *monitor.BuildInfo {
	if s.buildInfo == nil {
		return &monitor.BuildInfo{
			VEFVersion: constants.VEFVersion,
			AppVersion: "unknown",
			BuildTime:  "unknown",
			GitCommit:  "unknown",
		}
	}

	return s.buildInfo
}

// Init starts background goroutines to periodically sample cpu and process metrics.
func (s *DefaultService) Init(context.Context) error {
	samplerCtx, cancel := context.WithCancel(context.Background())
	s.samplerCancel = cancel
	s.samplerDone = make(chan struct{})

	// Start background sampling goroutine
	go func() {
		defer close(s.samplerDone)

		ticker := time.NewTicker(s.config.SampleInterval)
		defer ticker.Stop()

		// Perform initial sampling
		s.sampleCpu(samplerCtx)
		s.sampleProcess(samplerCtx)

		for {
			select {
			case <-samplerCtx.Done():
				return
			case <-ticker.C:
				// Sample both cpu and process metrics together
				s.sampleCpu(samplerCtx)
				s.sampleProcess(samplerCtx)
			}
		}
	}()

	return nil
}

// Close gracefully stops the background sampling goroutines.
func (s *DefaultService) Close() error {
	if s.samplerCancel != nil {
		s.samplerCancel()
	}

	if s.samplerDone != nil {
		<-s.samplerDone
	}

	return nil
}

// sampleCpu performs cpu sampling and stores the result in atomic cache.
func (s *DefaultService) sampleCpu(ctx context.Context) {
	cpuInfo, err := s.sampleCpuInfo(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}

		logger.Errorf("Failed to sample cpu info: %v", err)

		return
	}

	s.cpuCache.Store(cpuInfo)
}

// sampleCpuInfo is the actual cpu sampling implementation.
func (s *DefaultService) sampleCpuInfo(ctx context.Context) (*monitor.CpuInfo, error) {
	// Get cpu static info
	infoStat, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var cpuInfo monitor.CpuInfo
	if len(infoStat) > 0 {
		first := infoStat[0]
		cpuInfo.ModelName = first.ModelName
		cpuInfo.Mhz = first.Mhz
		cpuInfo.CacheSize = first.CacheSize
		cpuInfo.VendorId = first.VendorID
		cpuInfo.Family = first.Family
		cpuInfo.Model = first.Model
		cpuInfo.Stepping = first.Stepping
		cpuInfo.Microcode = first.Microcode
	}

	// Get core counts
	physicalCores, _ := cpu.CountsWithContext(ctx, false)
	logicalCores, _ := cpu.CountsWithContext(ctx, true)
	cpuInfo.PhysicalCores = physicalCores
	cpuInfo.LogicalCores = logicalCores

	// Get cpu usage (per-core with configured sampling window)
	perCorePercent, err := cpu.PercentWithContext(ctx, s.config.SampleDuration, true)
	if err == nil {
		cpuInfo.UsagePercent = perCorePercent
	}

	// Get total cpu usage (reuse last sample, no delay)
	totalPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(totalPercent) > 0 {
		cpuInfo.TotalPercent = totalPercent[0]
	}

	return &cpuInfo, nil
}

// sampleProcess performs process sampling and stores the result in atomic cache.
func (s *DefaultService) sampleProcess(ctx context.Context) {
	processInfo, err := s.sampleProcessInfo(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}

		logger.Errorf("Failed to sample process info: %v", err)

		return
	}

	s.processCache.Store(processInfo)
}

// sampleProcessInfo is the actual process sampling implementation.
func (s *DefaultService) sampleProcessInfo(ctx context.Context) (*monitor.ProcessInfo, error) {
	proc, err := process.NewProcessWithContext(ctx, int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	// Core metrics - must succeed
	cpuPercent, err := proc.PercentWithContext(ctx, s.config.SampleDuration)
	if err != nil {
		return nil, err
	}

	memPercent, err := proc.MemoryPercentWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Memory info - important but not critical
	memInfo, err := proc.MemoryInfoWithContext(ctx)
	if err != nil {
		logger.Warnf("Failed to get memory info: %v", err)
	}

	var memRss, memVms, memSwap uint64
	if memInfo != nil {
		memRss = memInfo.RSS
		memVms = memInfo.VMS
		memSwap = memInfo.Swap
	}

	// Optional fields - continue with zero values if failed
	name, _ := proc.NameWithContext(ctx)
	exe, _ := proc.ExeWithContext(ctx)
	cmdline, _ := proc.CmdlineWithContext(ctx)
	cwd, _ := proc.CwdWithContext(ctx)
	status, _ := proc.StatusWithContext(ctx)
	username, _ := proc.UsernameWithContext(ctx)
	createTime, _ := proc.CreateTimeWithContext(ctx)
	numThreads, _ := proc.NumThreadsWithContext(ctx)
	numFds, _ := proc.NumFDsWithContext(ctx)
	parentPid, _ := proc.PpidWithContext(ctx)

	var statusStr string
	if len(status) > 0 {
		statusStr = status[0]
	}

	return &monitor.ProcessInfo{
		Pid:           proc.Pid,
		ParentPid:     parentPid,
		Name:          name,
		Exe:           exe,
		Cmdline:       cmdline,
		Cwd:           cwd,
		Status:        statusStr,
		Username:      username,
		CreateTime:    createTime,
		NumThreads:    numThreads,
		NumFds:        numFds,
		CpuPercent:    cpuPercent,
		MemoryPercent: memPercent,
		MemoryRss:     memRss,
		MemoryVms:     memVms,
		MemorySwap:    memSwap,
	}, nil
}
