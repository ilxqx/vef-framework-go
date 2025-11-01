package monitor

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/contract"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/monitor"
)

var logger = log.Named("monitor")

// Module is the FX module for system monitoring functionality.
var Module = fx.Module(
	"vef:monitor",
	fx.Decorate(func(cfg *config.MonitorConfig) *config.MonitorConfig {
		// Merge with default config
		cfgToUse := DefaultConfig()
		if cfg != nil {
			if cfg.SampleInterval > 0 {
				cfgToUse.SampleInterval = cfg.SampleInterval
			}

			if cfg.SampleDuration > 0 {
				cfgToUse.SampleDuration = cfg.SampleDuration
			}
		}

		return &cfgToUse
	}),
	fx.Decorate(
		fx.Annotate(
			func(buildInfo *monitor.BuildInfo) *monitor.BuildInfo {
				// Populate framework version (override any user-provided value)
				if buildInfo != nil {
					buildInfo.VEFVersion = constants.VEFVersion
				} else {
					buildInfo = &monitor.BuildInfo{
						VEFVersion: constants.VEFVersion,
						AppVersion: "v0.0.0",
						BuildTime:  "2022-08-08 01:00:00",
						GitCommit:  "-",
					}
				}

				return buildInfo
			},
			fx.ParamTags(`optional:"true"`),
		),
	),
	fx.Provide(
		// Provide monitor service with lifecycle management
		fx.Annotate(
			NewService,
			fx.ParamTags(``, `optional:"true"`),
			fx.OnStart(func(ctx context.Context, svc monitor.Service) error {
				if initializer, ok := svc.(contract.Initializer); ok {
					if err := initializer.Init(ctx); err != nil {
						return fmt.Errorf("failed to initialize monitor service: %w", err)
					}
				}

				return nil
			}),
			fx.OnStop(func(svc monitor.Service) error {
				if closer, ok := svc.(io.Closer); ok {
					if err := closer.Close(); err != nil {
						return fmt.Errorf("failed to close monitor service: %w", err)
					}
				}

				return nil
			}),
		),
		// Provide monitor resource
		fx.Annotate(
			NewResource,
			fx.ResultTags(`group:"vef:api:resources"`),
		),
	),
)
