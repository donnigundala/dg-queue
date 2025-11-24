package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled jobs using cron syntax.
type Scheduler struct {
	cron    *cron.Cron
	manager *Manager
	entries map[string]cron.EntryID
	mu      sync.RWMutex
}

// NewScheduler creates a new scheduler.
func NewScheduler(manager *Manager) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		manager: manager,
		entries: make(map[string]cron.EntryID),
	}
}

// Schedule schedules a job using cron syntax.
// cronExpr: Cron expression (e.g., "*/5 * * * *" for every 5 minutes)
// name: Unique name for this scheduled job
// handler: Function to execute on schedule
func (s *Scheduler) Schedule(cronExpr, name string, handler ScheduleHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already scheduled
	if _, exists := s.entries[name]; exists {
		return fmt.Errorf("schedule '%s' already exists", name)
	}

	// Add to cron
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		if err := handler(); err != nil {
			// In production, you'd want to log this
			fmt.Printf("Scheduled job '%s' failed: %v\n", name, err)
		}
	})

	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	s.entries[name] = entryID
	return nil
}

// ScheduleJob schedules a job to be dispatched on a cron schedule.
// This is a convenience method that dispatches the job to the queue.
func (s *Scheduler) ScheduleJob(cronExpr, jobName string, payload interface{}) error {
	return s.Schedule(cronExpr, "schedule_"+jobName, func() error {
		_, err := s.manager.Dispatch(jobName, payload)
		return err
	})
}

// Remove removes a scheduled job.
func (s *Scheduler) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.entries[name]
	if !exists {
		return fmt.Errorf("schedule '%s' not found", name)
	}

	s.cron.Remove(entryID)
	delete(s.entries, name)
	return nil
}

// Start starts the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop stops the scheduler gracefully.
func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}

// Count returns the number of scheduled jobs.
func (s *Scheduler) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
