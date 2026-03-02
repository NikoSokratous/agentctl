package orchestrate

import (
	"context"
	"fmt"
	"time"
)

// Job represents a scheduled agent run.
type Job struct {
	AgentName string
	Goal      string
	CronExpr  string // e.g. "0 * * * *" for hourly
}

// Scheduler runs agents on a schedule.
type Scheduler struct {
	jobs  []Job
	runFn func(ctx context.Context, agent, goal string) error
}

// NewScheduler creates a scheduler.
func NewScheduler(runFn func(ctx context.Context, agent, goal string) error) *Scheduler {
	return &Scheduler{runFn: runFn}
}

// Add adds a job.
func (s *Scheduler) Add(job Job) {
	s.jobs = append(s.jobs, job)
}

// Run starts the scheduler (blocking).
func (s *Scheduler) Run(ctx context.Context) error {
	// Minimal implementation: run jobs on interval for MVP.
	// Full cron parsing can be added with robfig/cron.
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, j := range s.jobs {
				if err := s.runFn(ctx, j.AgentName, j.Goal); err != nil {
					fmt.Printf("scheduler job %s: %v\n", j.AgentName, err)
				}
			}
		}
	}
}
