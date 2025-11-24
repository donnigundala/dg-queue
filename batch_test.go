package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/donnigundala/dg-queue/drivers/memory"
)

func TestBatch_DispatchBatch(t *testing.T) {
	manager := New(DefaultConfig())
	manager.SetDriver(memory.NewDriver())
	batch := NewBatch(manager)

	// Register worker
	processed := 0
	var mu sync.Mutex
	manager.Worker("batch-job", 5, func(job *Job) error {
		mu.Lock()
		processed++
		mu.Unlock()
		return nil
	})

	items := []interface{}{
		map[string]string{"id": "1"},
		map[string]string{"id": "2"},
		map[string]string{"id": "3"},
	}

	status, err := batch.DispatchBatch("batch-job", items, DefaultBatchConfig())
	if err != nil {
		t.Fatalf("Failed to dispatch batch: %v", err)
	}

	if status.Total != 3 {
		t.Errorf("Expected total 3, got %d", status.Total)
	}

	// Wait for batch to complete
	time.Sleep(100 * time.Millisecond)

	if !status.IsComplete() {
		t.Error("Expected batch to be complete")
	}

	if status.Processed != 3 {
		t.Errorf("Expected 3 processed, got %d", status.Processed)
	}
}

func TestBatch_EmptyItems(t *testing.T) {
	manager := New(DefaultConfig())
	batch := NewBatch(manager)

	items := []interface{}{}
	_, err := batch.DispatchBatch("test", items, DefaultBatchConfig())
	if err == nil {
		t.Error("Expected error for empty items")
	}
}

func TestBatch_Chunking(t *testing.T) {
	manager := New(DefaultConfig())
	manager.SetDriver(memory.NewDriver())
	batch := NewBatch(manager)

	// Create 250 items
	items := make([]interface{}, 250)
	for i := 0; i < 250; i++ {
		items[i] = i
	}

	config := BatchConfig{
		ChunkSize:       100,
		ContinueOnError: true,
	}

	status, err := batch.DispatchBatch("test", items, config)
	if err != nil {
		t.Fatalf("Failed to dispatch batch: %v", err)
	}

	if status.Total != 250 {
		t.Errorf("Expected total 250, got %d", status.Total)
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	if status.Processed != 250 {
		t.Errorf("Expected 250 processed, got %d", status.Processed)
	}
}

func TestBatch_ProgressCallback(t *testing.T) {
	manager := New(DefaultConfig())
	manager.SetDriver(memory.NewDriver())
	batch := NewBatch(manager)

	progressCalls := 0
	var mu sync.Mutex

	items := []interface{}{1, 2, 3, 4, 5}

	config := BatchConfig{
		ChunkSize: 100,
		OnProgress: func(processed, total int) {
			mu.Lock()
			progressCalls++
			mu.Unlock()
		},
		ContinueOnError: true,
	}

	status, err := batch.DispatchBatch("test", items, config)
	if err != nil {
		t.Fatalf("Failed to dispatch batch: %v", err)
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if progressCalls != 5 {
		t.Errorf("Expected 5 progress callbacks, got %d", progressCalls)
	}

	if status.Progress() != 100 {
		t.Errorf("Expected progress 100%%, got %.2f%%", status.Progress())
	}
}

func TestBatch_ErrorHandling(t *testing.T) {
	manager := New(DefaultConfig())
	manager.SetDriver(memory.NewDriver())
	batch := NewBatch(manager)

	errorCount := 0
	var mu sync.Mutex

	// Simulate errors by using non-registered job
	items := []interface{}{1, 2, 3}

	config := BatchConfig{
		ChunkSize: 100,
		OnError: func(item interface{}, err error) {
			mu.Lock()
			errorCount++
			mu.Unlock()
		},
		ContinueOnError: true,
	}

	status, err := batch.DispatchBatch("nonexistent-job", items, config)
	if err != nil {
		t.Fatalf("Failed to dispatch batch: %v", err)
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if errorCount != 3 {
		t.Errorf("Expected 3 errors, got %d", errorCount)
	}

	if status.Failed != 3 {
		t.Errorf("Expected 3 failed, got %d", status.Failed)
	}
}

func TestBatch_Map(t *testing.T) {
	manager := New(DefaultConfig())
	manager.SetDriver(memory.NewDriver())
	batch := NewBatch(manager)

	items := []interface{}{1, 2, 3}

	// Mapper function
	mapper := func(item interface{}) (interface{}, error) {
		num := item.(int)
		return map[string]int{"value": num * 2}, nil
	}

	status, err := batch.Map("test", items, mapper, DefaultBatchConfig())
	if err != nil {
		t.Fatalf("Failed to map batch: %v", err)
	}

	if status.Total != 3 {
		t.Errorf("Expected total 3, got %d", status.Total)
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	if status.Processed != 3 {
		t.Errorf("Expected 3 processed, got %d", status.Processed)
	}
}

func TestBatch_MapWithError(t *testing.T) {
	manager := New(DefaultConfig())
	batch := NewBatch(manager)

	items := []interface{}{1, 2, 3}

	// Mapper that fails
	mapper := func(item interface{}) (interface{}, error) {
		return nil, fmt.Errorf("mapping error")
	}

	config := BatchConfig{
		ChunkSize:       100,
		ContinueOnError: false,
	}

	_, err := batch.Map("test", items, mapper, config)
	if err == nil {
		t.Error("Expected error when mapping fails and ContinueOnError is false")
	}
}

func TestBatchStatus_Progress(t *testing.T) {
	status := &BatchStatus{
		Total:     100,
		Processed: 50,
	}

	progress := status.Progress()
	if progress != 50.0 {
		t.Errorf("Expected progress 50%%, got %.2f%%", progress)
	}

	// Zero total
	status.Total = 0
	progress = status.Progress()
	if progress != 0 {
		t.Errorf("Expected progress 0%% for zero total, got %.2f%%", progress)
	}
}
