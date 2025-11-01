package monitor_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/contract"
	imonitor "github.com/ilxqx/vef-framework-go/internal/monitor"
	"github.com/ilxqx/vef-framework-go/monitor"
)

// MonitorServiceTestSuite tests the monitor service implementation.
type MonitorServiceTestSuite struct {
	suite.Suite

	ctx                         context.Context
	serviceWithCustomBuildInfo  monitor.Service
	serviceWithDefaultBuildInfo monitor.Service
}

func (suite *MonitorServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	cfg := &config.MonitorConfig{
		SampleInterval: 100 * time.Millisecond,
		SampleDuration: 50 * time.Millisecond,
	}

	buildInfo := &monitor.BuildInfo{
		VEFVersion: constants.VEFVersion,
		AppVersion: "v1.0.0",
		BuildTime:  "2024-01-01T00:00:00Z",
		GitCommit:  "abc123def456",
	}

	suite.serviceWithCustomBuildInfo = imonitor.NewService(cfg, buildInfo)
	if initializer, ok := suite.serviceWithCustomBuildInfo.(contract.Initializer); ok {
		err := initializer.Init(suite.ctx)
		suite.Require().NoError(err)
	}

	suite.serviceWithDefaultBuildInfo = imonitor.NewService(cfg, nil)
	if initializer, ok := suite.serviceWithDefaultBuildInfo.(contract.Initializer); ok {
		err := initializer.Init(suite.ctx)
		suite.Require().NoError(err)
	}

	time.Sleep(100 * time.Millisecond)
}

func (suite *MonitorServiceTestSuite) TearDownSuite() {
	if suite.serviceWithCustomBuildInfo != nil {
		if closer, ok := suite.serviceWithCustomBuildInfo.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				suite.T().Logf("failed to close monitor service: %v", err)
			}
		}
	}

	if suite.serviceWithDefaultBuildInfo != nil {
		if closer, ok := suite.serviceWithDefaultBuildInfo.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				suite.T().Logf("failed to close monitor service: %v", err)
			}
		}
	}
}

// TestOverview tests the system overview functionality.
func (suite *MonitorServiceTestSuite) TestOverview() {
	suite.T().Log("Testing Overview method")

	suite.Run("WithCustomBuildInfo", func() {
		overview, err := suite.serviceWithCustomBuildInfo.Overview(suite.ctx)
		suite.NoError(err, "Overview should not return error")
		suite.NotNil(overview, "Overview should not be nil")

		suite.NotNil(overview.Host, "Host info should be present")
		suite.NotNil(overview.Cpu, "Cpu info should be present")
		suite.NotNil(overview.Memory, "Memory info should be present")
		suite.NotNil(overview.Disk, "Disk info should be present")
		suite.NotNil(overview.Network, "Network info should be present")
		suite.NotNil(overview.Process, "Process info should be present")
		suite.NotNil(overview.Load, "Load info should be present")
		suite.NotNil(overview.Build, "Build info should be present")

		suite.Equal("v1.0.0", overview.Build.AppVersion, "AppVersion should match")
		suite.NotEmpty(overview.Build.VEFVersion, "VEFVersion should be populated")
		suite.Equal("2024-01-01T00:00:00Z", overview.Build.BuildTime, "BuildTime should match")
		suite.Equal("abc123def456", overview.Build.GitCommit, "GitCommit should match")
	})

	suite.Run("WithDefaultBuildInfo", func() {
		overview, err := suite.serviceWithDefaultBuildInfo.Overview(suite.ctx)
		suite.NoError(err, "Overview should not return error")
		suite.NotNil(overview, "Overview should not be nil")

		suite.NotNil(overview.Build, "Build info should be present")
		suite.Equal("unknown", overview.Build.AppVersion, "Should have default app version")
		suite.NotEmpty(overview.Build.VEFVersion, "VEFVersion should be populated")
		suite.Equal("unknown", overview.Build.BuildTime, "Should have default build time")
		suite.Equal("unknown", overview.Build.GitCommit, "Should have default git commit")
	})
}

// TestCpu tests CPU information retrieval.
func (suite *MonitorServiceTestSuite) TestCpu() {
	suite.T().Log("Testing Cpu method")

	suite.Run("Success", func() {
		cpuInfo, err := suite.serviceWithCustomBuildInfo.Cpu(suite.ctx)
		suite.NoError(err, "Cpu should not return error")
		suite.NotNil(cpuInfo, "CpuInfo should not be nil")

		suite.Greater(cpuInfo.PhysicalCores, 0, "Should have at least 1 physical core")
		suite.Greater(cpuInfo.LogicalCores, 0, "Should have at least 1 logical core")
		suite.GreaterOrEqual(cpuInfo.LogicalCores, cpuInfo.PhysicalCores, "Logical cores should be >= physical cores")

		suite.NotNil(cpuInfo.UsagePercent, "Per-core usage should be present")
		suite.GreaterOrEqual(cpuInfo.TotalPercent, 0.0, "Total CPU percent should be >= 0")
		suite.LessOrEqual(cpuInfo.TotalPercent, 100.0*float64(cpuInfo.LogicalCores), "Total CPU percent should be reasonable")
	})
}

// TestMemory tests memory information retrieval.
func (suite *MonitorServiceTestSuite) TestMemory() {
	suite.T().Log("Testing Memory method")

	suite.Run("Success", func() {
		memInfo, err := suite.serviceWithCustomBuildInfo.Memory(suite.ctx)
		suite.NoError(err, "Memory should not return error")
		suite.NotNil(memInfo, "MemoryInfo should not be nil")
		suite.NotNil(memInfo.Virtual, "Virtual memory should be present")

		suite.Greater(memInfo.Virtual.Total, uint64(0), "Total memory should be > 0")
		suite.LessOrEqual(memInfo.Virtual.Used, memInfo.Virtual.Total, "Used memory should be <= total")
		suite.GreaterOrEqual(memInfo.Virtual.UsedPercent, 0.0, "Used percent should be >= 0")
		suite.LessOrEqual(memInfo.Virtual.UsedPercent, 100.0, "Used percent should be <= 100")

		if memInfo.Swap != nil {
			suite.GreaterOrEqual(memInfo.Swap.Total, uint64(0), "Swap total should be >= 0")
		}
	})
}

// TestDisk tests disk information retrieval.
func (suite *MonitorServiceTestSuite) TestDisk() {
	suite.T().Log("Testing Disk method")

	suite.Run("Success", func() {
		diskInfo, err := suite.serviceWithCustomBuildInfo.Disk(suite.ctx)
		suite.NoError(err, "Disk should not return error")
		suite.NotNil(diskInfo, "DiskInfo should not be nil")

		suite.NotEmpty(diskInfo.Partitions, "Should have at least one partition")

		for _, part := range diskInfo.Partitions {
			suite.NotEmpty(part.MountPoint, "MountPoint should not be empty")

			if part.Total > 0 {
				suite.LessOrEqual(part.Used, part.Total, "Used should be <= total")
			}
		}
	})
}

// TestNetwork tests network information retrieval.
func (suite *MonitorServiceTestSuite) TestNetwork() {
	suite.T().Log("Testing Network method")

	suite.Run("Success", func() {
		netInfo, err := suite.serviceWithCustomBuildInfo.Network(suite.ctx)
		suite.NoError(err, "Network should not return error")
		suite.NotNil(netInfo, "NetworkInfo should not be nil")

		suite.NotEmpty(netInfo.Interfaces, "Should have at least one network interface")

		for _, iface := range netInfo.Interfaces {
			suite.NotEmpty(iface.Name, "Interface name should not be empty")
		}

		suite.NotNil(netInfo.IoCounters, "IO counters should be present")
	})
}

// TestHost tests host information retrieval.
func (suite *MonitorServiceTestSuite) TestHost() {
	suite.T().Log("Testing Host method")

	suite.Run("FirstCall", func() {
		hostInfo1, err := suite.serviceWithCustomBuildInfo.Host(suite.ctx)
		suite.NoError(err, "Host should not return error")
		suite.NotNil(hostInfo1, "HostInfo should not be nil")

		suite.NotEmpty(hostInfo1.Hostname, "Hostname should not be empty")
		suite.NotEmpty(hostInfo1.Os, "OS should not be empty")
		suite.NotEmpty(hostInfo1.Platform, "Platform should not be empty")
	})

	suite.Run("CachedCall", func() {
		hostInfo1, err := suite.serviceWithCustomBuildInfo.Host(suite.ctx)
		suite.NoError(err, "First call should not return error")

		hostInfo2, err := suite.serviceWithCustomBuildInfo.Host(suite.ctx)
		suite.NoError(err, "Second call should not return error")

		suite.Equal(hostInfo1.Hostname, hostInfo2.Hostname, "Hostname should be consistent")
		suite.Equal(hostInfo1.Os, hostInfo2.Os, "OS should be consistent")
		suite.Equal(hostInfo1.Platform, hostInfo2.Platform, "Platform should be consistent")
	})
}

// TestProcess tests process information retrieval.
func (suite *MonitorServiceTestSuite) TestProcess() {
	suite.T().Log("Testing Process method")

	suite.Run("Success", func() {
		procInfo, err := suite.serviceWithCustomBuildInfo.Process(suite.ctx)
		suite.NoError(err, "Process should not return error")
		suite.NotNil(procInfo, "ProcessInfo should not be nil")

		suite.Greater(procInfo.Pid, int32(0), "PID should be > 0")
		suite.NotEmpty(procInfo.Name, "Process name should not be empty")
		suite.GreaterOrEqual(procInfo.CpuPercent, 0.0, "CPU percent should be >= 0")
		suite.GreaterOrEqual(procInfo.MemoryPercent, float32(0.0), "Memory percent should be >= 0")
		suite.Greater(procInfo.MemoryRss, uint64(0), "Memory RSS should be > 0")
	})
}

// TestLoad tests system load averages retrieval.
func (suite *MonitorServiceTestSuite) TestLoad() {
	suite.T().Log("Testing Load method")

	suite.Run("Success", func() {
		loadInfo, err := suite.serviceWithCustomBuildInfo.Load(suite.ctx)
		suite.NoError(err, "Load should not return error")
		suite.NotNil(loadInfo, "LoadInfo should not be nil")

		suite.GreaterOrEqual(loadInfo.Load1, 0.0, "Load1 should be >= 0")
		suite.GreaterOrEqual(loadInfo.Load5, 0.0, "Load5 should be >= 0")
		suite.GreaterOrEqual(loadInfo.Load15, 0.0, "Load15 should be >= 0")
	})
}

// TestBuildInfo tests build information retrieval.
func (suite *MonitorServiceTestSuite) TestBuildInfo() {
	suite.T().Log("Testing BuildInfo method")

	suite.Run("WithCustom", func() {
		buildInfo := suite.serviceWithCustomBuildInfo.BuildInfo()
		suite.NotNil(buildInfo, "BuildInfo should not be nil")

		suite.Equal("v1.0.0", buildInfo.AppVersion, "AppVersion should match")
		suite.NotEmpty(buildInfo.VEFVersion, "VEFVersion should be populated")
		suite.Equal("2024-01-01T00:00:00Z", buildInfo.BuildTime, "BuildTime should match")
		suite.Equal("abc123def456", buildInfo.GitCommit, "GitCommit should match")
	})

	suite.Run("WithDefault", func() {
		buildInfo := suite.serviceWithDefaultBuildInfo.BuildInfo()
		suite.NotNil(buildInfo, "BuildInfo should not be nil")

		suite.Equal("unknown", buildInfo.AppVersion, "Should have default app version")
		suite.NotEmpty(buildInfo.VEFVersion, "VEFVersion should be populated")
		suite.Equal("unknown", buildInfo.BuildTime, "Should have default build time")
		suite.Equal("unknown", buildInfo.GitCommit, "Should have default git commit")
	})
}

func TestMonitorServiceSuite(t *testing.T) {
	suite.Run(t, new(MonitorServiceTestSuite))
}
