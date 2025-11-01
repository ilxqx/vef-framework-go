package monitor

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/monitor"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// NewResource creates a new monitor resource with the provided service.
func NewResource(service monitor.Service) api.Resource {
	// In test environment, make all actions public (no authentication required)
	isPublic := testhelpers.IsTestEnv()

	return &Resource{
		service: service,
		Resource: api.NewResource(
			"sys/monitor",
			api.WithApis(
				api.Spec{Action: "get_overview", Public: isPublic},
				api.Spec{Action: "get_cpu", Public: isPublic},
				api.Spec{Action: "get_memory", Public: isPublic},
				api.Spec{Action: "get_disk", Public: isPublic},
				api.Spec{Action: "get_network", Public: isPublic},
				api.Spec{Action: "get_host", Public: isPublic},
				api.Spec{Action: "get_process", Public: isPublic},
				api.Spec{Action: "get_load", Public: isPublic},
				api.Spec{Action: "get_build_info", Public: isPublic},
			),
		),
	}
}

// Resource handles system monitoring-related API endpoints.
type Resource struct {
	api.Resource

	service monitor.Service
}

// GetOverview returns a comprehensive system overview.
func (r *Resource) GetOverview(ctx fiber.Ctx) error {
	overview, err := r.service.Overview(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(overview).Response(ctx)
}

// GetCpu returns detailed cpu information.
func (r *Resource) GetCpu(ctx fiber.Ctx) error {
	cpuInfo, err := r.service.Cpu(ctx.Context())
	if err != nil {
		return result.Err(
			i18n.T("monitor_not_ready"),
			result.WithCode(result.ErrCodeMonitorNotReady),
		)
	}

	return result.Ok(cpuInfo).Response(ctx)
}

// GetMemory returns memory usage information.
func (r *Resource) GetMemory(ctx fiber.Ctx) error {
	memInfo, err := r.service.Memory(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(memInfo).Response(ctx)
}

// GetDisk returns disk usage and partition information.
func (r *Resource) GetDisk(ctx fiber.Ctx) error {
	diskInfo, err := r.service.Disk(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(diskInfo).Response(ctx)
}

// GetNetwork returns network interface and I/O statistics.
func (r *Resource) GetNetwork(ctx fiber.Ctx) error {
	netInfo, err := r.service.Network(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(netInfo).Response(ctx)
}

// GetHost returns static host information.
func (r *Resource) GetHost(ctx fiber.Ctx) error {
	hostInfo, err := r.service.Host(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(hostInfo).Response(ctx)
}

// GetProcess returns information about the current process.
func (r *Resource) GetProcess(ctx fiber.Ctx) error {
	procInfo, err := r.service.Process(ctx.Context())
	if err != nil {
		return result.Err(
			i18n.T("monitor_not_ready"),
			result.WithCode(result.ErrCodeMonitorNotReady),
		)
	}

	return result.Ok(procInfo).Response(ctx)
}

// GetLoad returns system load averages.
func (r *Resource) GetLoad(ctx fiber.Ctx) error {
	loadInfo, err := r.service.Load(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(loadInfo).Response(ctx)
}

// GetBuildInfo returns application build information.
func (r *Resource) GetBuildInfo(ctx fiber.Ctx) error {
	return result.Ok(r.service.BuildInfo()).Response(ctx)
}
