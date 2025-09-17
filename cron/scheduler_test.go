package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestScheduler creates a gocron.Scheduler instance for testing
// Similar to the production configuration but simplified for tests
func createTestScheduler() (gocron.Scheduler, error) {
	return gocron.NewScheduler(
		gocron.WithLocation(time.Local),
		gocron.WithStopTimeout(5*time.Second), // Shorter timeout for tests
	)
}

func TestNewScheduler(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)
	assert.NotNil(t, scheduler)
}

func TestScheduler_NewJob_OneTime(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	testFunc := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Create a one-time job that runs immediately
	jobDef := NewOneTimeJob(nil,
		WithName("test-one-time"),
		WithTags("test", "one-time"),
		WithTask(testFunc),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-one-time", job.Name())
	assert.Contains(t, job.Tags(), "test")
	assert.Contains(t, job.Tags(), "one-time")

	// Start scheduler and wait for execution
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
}

func TestScheduler_NewJob_Duration(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	testFunc := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Create a duration-based job with limited runs
	jobDef := NewDurationJob(50*time.Millisecond,
		WithName("test-duration"),
		WithTags("test", "duration"),
		WithLimitedRuns(3),
		WithTask(testFunc),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-duration", job.Name())

	// Start scheduler and wait for multiple executions
	scheduler.Start()
	time.Sleep(200 * time.Millisecond)

	// Should have executed 3 times due to limit
	assert.Equal(t, int32(3), atomic.LoadInt32(&executed))
}

func TestScheduler_NewJob_Cron(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	testFunc := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Create a cron job that runs every second
	jobDef := NewCronJob("* * * * * *", true, // every second with seconds field
		WithName("test-cron"),
		WithTags("test", "cron"),
		WithLimitedRuns(2),
		WithTask(testFunc),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-cron", job.Name())

	// Start scheduler and wait for executions
	scheduler.Start()
	time.Sleep(2500 * time.Millisecond) // Wait for 2+ seconds

	// Should have executed 2 times due to limit
	assert.Equal(t, int32(2), atomic.LoadInt32(&executed))
}

func TestScheduler_NewJob_DurationRandom(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	testFunc := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Create a random duration job
	jobDef := NewDurationRandomJob(10*time.Millisecond, 50*time.Millisecond,
		WithName("test-random"),
		WithTags("test", "random"),
		WithLimitedRuns(2),
		WithTask(testFunc),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-random", job.Name())

	// Start scheduler and wait for executions
	scheduler.Start()
	time.Sleep(200 * time.Millisecond)

	// Should have executed at least once, possibly twice
	executions := atomic.LoadInt32(&executed)
	assert.GreaterOrEqual(t, executions, int32(1))
	assert.LessOrEqual(t, executions, int32(2))
}

func TestScheduler_Jobs(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	// Initially no jobs
	jobs := scheduler.Jobs()
	assert.Len(t, jobs, 0)

	// Add a job
	jobDef := NewOneTimeJob(nil,
		WithName("test-job"),
		WithTask(func() {}),
	)

	_, err = scheduler.NewJob(jobDef)
	require.NoError(t, err)

	// Should now have one job
	jobs = scheduler.Jobs()
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test-job", jobs[0].Name())
}

func TestScheduler_RemoveJob(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	// Add a job
	jobDef := NewOneTimeJob(nil,
		WithName("test-job"),
		WithTask(func() {}),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)

	// Verify job exists
	jobs := scheduler.Jobs()
	assert.Len(t, jobs, 1)

	// Remove job by ID
	err = scheduler.RemoveJob(job.Id())
	require.NoError(t, err)

	// Verify job is removed
	jobs = scheduler.Jobs()
	assert.Len(t, jobs, 0)
}

func TestScheduler_RemoveByTags(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	// Add jobs with different tags
	job1Def := NewOneTimeJob(nil,
		WithName("job1"),
		WithTags("group1", "test"),
		WithTask(func() {}),
	)

	job2Def := NewOneTimeJob(nil,
		WithName("job2"),
		WithTags("group2", "test"),
		WithTask(func() {}),
	)

	job3Def := NewOneTimeJob(nil,
		WithName("job3"),
		WithTags("group1"),
		WithTask(func() {}),
	)

	_, err = scheduler.NewJob(job1Def)
	require.NoError(t, err)
	_, err = scheduler.NewJob(job2Def)
	require.NoError(t, err)
	_, err = scheduler.NewJob(job3Def)
	require.NoError(t, err)

	// Verify all jobs exist
	jobs := scheduler.Jobs()
	assert.Len(t, jobs, 3)

	// Remove jobs with "test" tag
	scheduler.RemoveByTags("test")

	// Should have removed job1 and job2, leaving job3
	jobs = scheduler.Jobs()
	assert.Len(t, jobs, 1)
	assert.Equal(t, "job3", jobs[0].Name())
}

func TestScheduler_UpdateJob(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed1, executed2 int32
	testFunc1 := func() {
		atomic.AddInt32(&executed1, 1)
	}
	testFunc2 := func() {
		atomic.AddInt32(&executed2, 1)
	}

	// Create initial job
	jobDef1 := NewOneTimeJob(nil,
		WithName("test-job"),
		WithTask(testFunc1),
	)

	job, err := scheduler.NewJob(jobDef1)
	require.NoError(t, err)
	originalID := job.Id()

	// Update with new definition
	jobDef2 := NewOneTimeJob(nil,
		WithName("updated-job"),
		WithTask(testFunc2),
	)

	updatedJob, err := scheduler.Update(originalID, jobDef2)
	require.NoError(t, err)
	assert.Equal(t, "updated-job", updatedJob.Name())

	// Start scheduler
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)

	// Only the updated function should have executed
	assert.Equal(t, int32(0), atomic.LoadInt32(&executed1))
	assert.Equal(t, int32(1), atomic.LoadInt32(&executed2))
}

func TestScheduler_WithContext(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	ctx, cancel := context.WithCancel(context.Background())

	testFunc := func(jobCtx context.Context) {
		select {
		case <-jobCtx.Done():
			// Context was cancelled
			return
		case <-time.After(100 * time.Millisecond):
			atomic.AddInt32(&executed, 1)
		}
	}

	// Create job with context
	jobDef := NewDurationJob(50*time.Millisecond,
		WithName("test-context"),
		WithContext(ctx),
		WithTask(testFunc),
	)

	_, err = scheduler.NewJob(jobDef)
	require.NoError(t, err)

	// Start scheduler
	scheduler.Start()
	time.Sleep(25 * time.Millisecond) // Let it start

	// Cancel context before job can complete
	cancel()
	time.Sleep(200 * time.Millisecond)

	// Job should not have executed due to context cancellation
	assert.Equal(t, int32(0), atomic.LoadInt32(&executed))
}

func TestScheduler_StopJobs(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	testFunc := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Create a recurring job
	jobDef := NewDurationJob(50*time.Millisecond,
		WithName("test-stop"),
		WithTask(testFunc),
	)

	_, err = scheduler.NewJob(jobDef)
	require.NoError(t, err)

	// Start scheduler and let it run
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)

	// Stop jobs
	err = scheduler.StopJobs()
	require.NoError(t, err)

	// Record executions after stop
	executedAfterStop := atomic.LoadInt32(&executed)

	// Wait and verify no more executions
	time.Sleep(150 * time.Millisecond)
	finalExecuted := atomic.LoadInt32(&executed)

	assert.Equal(t, executedAfterStop, finalExecuted, "Jobs should not execute after being stopped")
}

func TestJob_RunNow(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	var executed int32
	testFunc := func() {
		atomic.AddInt32(&executed, 1)
	}

	// Create a job that would normally run later
	futureTime := time.Now().Add(1 * time.Hour)
	jobDef := NewOneTimeJob([]time.Time{futureTime},
		WithName("test-run-now"),
		WithTask(testFunc),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)

	scheduler.Start()

	// Job shouldn't have executed yet
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&executed))

	// Run now
	err = job.RunNow()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
}

func TestJob_NextRuns(t *testing.T) {
	gocronScheduler, err := createTestScheduler()
	require.NoError(t, err)
	defer gocronScheduler.Shutdown()

	scheduler := NewScheduler(gocronScheduler)

	// Create a job that runs every minute
	jobDef := NewCronJob("0 * * * * *", true, // every minute at second 0
		WithName("test-next-runs"),
		WithTask(func() {}),
	)

	job, err := scheduler.NewJob(jobDef)
	require.NoError(t, err)

	// Start scheduler so the job can be scheduled
	scheduler.Start()
	time.Sleep(10 * time.Millisecond) // Give it a moment to start

	// Get next 3 run times
	nextRuns, err := job.NextRuns(3)
	require.NoError(t, err)
	assert.Len(t, nextRuns, 3)

	// Verify times are in ascending order and roughly 1 minute apart
	for i := 1; i < len(nextRuns); i++ {
		diff := nextRuns[i].Sub(nextRuns[i-1])
		assert.InDelta(t, time.Minute, diff, float64(time.Second), "Next runs should be ~1 minute apart")
	}
}
