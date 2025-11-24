package queue

import (
	"fmt"
	"time"
)

// Batch provides batch processing capabilities.
type Batch struct {
	manager *Manager
}

// NewBatch creates a new batch processor.
func NewBatch(manager *Manager) *Batch {
	return &Batch{
		manager: manager,
	}
}

// DispatchBatch dispatches multiple jobs in batches.
func (b *Batch) DispatchBatch(name string, items []interface{}, config BatchConfig) (*BatchStatus, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("items cannot be empty")
	}

	status := &BatchStatus{
		Total:      len(items),
		Processed:  0,
		Failed:     0,
		StartedAt:  time.Now(),
		InProgress: true,
	}

	// Process items in chunks
	chunkSize := config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 100 // Default chunk size
	}

	go func() {
		defer func() {
			status.InProgress = false
			status.CompletedAt = time.Now()
		}()

		for i := 0; i < len(items); i += chunkSize {
			end := i + chunkSize
			if end > len(items) {
				end = len(items)
			}

			chunk := items[i:end]

			// Process chunk
			for _, item := range chunk {
				job, err := b.manager.Dispatch(name, item)
				if err != nil {
					status.Failed++
					if config.OnError != nil {
						config.OnError(item, err)
					}
					if !config.ContinueOnError {
						return
					}
					continue
				}

				status.Processed++
				status.JobIDs = append(status.JobIDs, job.ID)

				// Progress callback
				if config.OnProgress != nil {
					config.OnProgress(status.Processed, status.Total)
				}
			}

			// Rate limiting
			if config.RateLimit > 0 && i+chunkSize < len(items) {
				delay := time.Duration(chunkSize) * time.Second / time.Duration(config.RateLimit)
				time.Sleep(delay)
			}
		}
	}()

	return status, nil
}

// Map applies a mapper function to each item and dispatches the result.
func (b *Batch) Map(name string, items []interface{}, mapper BatchMapper, config BatchConfig) (*BatchStatus, error) {
	if mapper == nil {
		return nil, fmt.Errorf("mapper cannot be nil")
	}

	// Map items
	mappedItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		mapped, err := mapper(item)
		if err != nil {
			if config.OnError != nil {
				config.OnError(item, err)
			}
			if !config.ContinueOnError {
				return nil, err
			}
			continue
		}
		mappedItems = append(mappedItems, mapped)
	}

	return b.DispatchBatch(name, mappedItems, config)
}

// BatchStatus represents the status of a batch operation.
type BatchStatus struct {
	Total       int
	Processed   int
	Failed      int
	JobIDs      []string
	StartedAt   time.Time
	CompletedAt time.Time
	InProgress  bool
}

// Progress returns the progress percentage.
func (bs *BatchStatus) Progress() float64 {
	if bs.Total == 0 {
		return 0
	}
	return float64(bs.Processed) / float64(bs.Total) * 100
}

// IsComplete returns true if the batch is complete.
func (bs *BatchStatus) IsComplete() bool {
	return !bs.InProgress
}
