package queue

import (
	"sync"
	"testing"
	"time"
)

func TestScheduler_Schedule(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	executed := false
	var mu sync.Mutex

	// Every minute
	err := scheduler.Schedule("* * * * *", "test-every-minute", func() error {
		mu.Lock()
		executed = true
		mu.Unlock()
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to schedule: %v", err)
	}

	scheduler.Start()

	// Wait for execution (1 minute + buffer)
	time.Sleep(65 * time.Second)

	mu.Lock()
	if !executed {
		t.Error("Expected scheduled job to execute")
	}
	mu.Unlock()
}

func TestScheduler_InvalidCron(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	err := scheduler.Schedule("invalid cron", "test", func() error {
		return nil
	})

	if err == nil {
		t.Error("Expected error for invalid cron expression")
	}
}

func TestScheduler_DuplicateName(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	// Add first schedule
	err := scheduler.Schedule("*/5 * * * *", "duplicate", func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to add first schedule: %v", err)
	}

	// Try to add duplicate
	err = scheduler.Schedule("*/10 * * * *", "duplicate", func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected error for duplicate schedule name")
	}
}

func TestScheduler_Remove(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	// Add schedule
	scheduler.Schedule("*/5 * * * *", "removable", func() error {
		return nil
	})

	if scheduler.Count() != 1 {
		t.Errorf("Expected 1 scheduled job, got %d", scheduler.Count())
	}

	// Remove
	err := scheduler.Remove("removable")
	if err != nil {
		t.Fatalf("Failed to remove schedule: %v", err)
	}

	if scheduler.Count() != 0 {
		t.Errorf("Expected 0 scheduled jobs after removal, got %d", scheduler.Count())
	}
}

func TestScheduler_RemoveNonExistent(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	err := scheduler.Remove("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent schedule")
	}
}

func TestScheduler_Count(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	if scheduler.Count() != 0 {
		t.Errorf("Expected 0 schedules, got %d", scheduler.Count())
	}

	scheduler.Schedule("*/5 * * * *", "test1", func() error { return nil })
	scheduler.Schedule("*/10 * * * *", "test2", func() error { return nil })

	if scheduler.Count() != 2 {
		t.Errorf("Expected 2 schedules, got %d", scheduler.Count())
	}
}

func TestScheduler_ScheduleJob(t *testing.T) {
	manager := New(DefaultConfig())
	scheduler := NewScheduler(manager)
	defer scheduler.Stop()

	// This tests the convenience method
	err := scheduler.ScheduleJob("*/5 * * * *", "test-job", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Failed to schedule job: %v", err)
	}

	if scheduler.Count() != 1 {
		t.Errorf("Expected 1 scheduled job, got %d", scheduler.Count())
	}
}
