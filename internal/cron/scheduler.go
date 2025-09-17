package cron

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"go.uber.org/fx"
)

var logger = log.Named("cron")

// newScheduler creates a new gocron scheduler with optimal configuration for production use.
// It sets up logging, monitoring, concurrency limits, and lifecycle management.
func newScheduler(lc fx.Lifecycle) (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local),
		gocron.WithStopTimeout(30*time.Second),
		gocron.WithLogger(newCronLogger()),
		gocron.WithMonitorStatus(newJobMonitor()),
		gocron.WithLimitConcurrentJobs(1000, gocron.LimitModeWait),
		// gocron.WithGlobalJobOptions(
		// 	gocron.WithSingletonMode(gocron.LimitModeWait),
		// ),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}

	lc.Append(
		fx.StartStopHook(
			func() {
				scheduler.Start()
				logger.Info("cron scheduler started")
			},
			func() error {
				if err := scheduler.Shutdown(); err != nil {
					return fmt.Errorf("failed to stop scheduler: %w", err)
				}

				logger.Info("cron scheduler stopped")
				return nil
			},
		),
	)

	return scheduler, nil
}
