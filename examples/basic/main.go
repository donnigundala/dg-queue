package main

import (
	"context"
	"fmt"
	"log"
	"time"

	queue "github.com/donnigundala/dg-queue"
	"github.com/donnigundala/dg-queue/drivers/memory"
)

type EmailJob struct {
	To      string
	Subject string
	Body    string
}

func main() {
	// Create queue with memory driver
	q := queue.New(queue.DefaultConfig())
	q.SetDriver(memory.NewDriver())

	// Register worker
	q.Worker("send-email", 5, func(job *queue.Job) error {
		email := job.Payload.(map[string]interface{})
		fmt.Printf("[Worker] Sending email to: %s\n", email["to"])
		fmt.Printf("[Worker] Subject: %s\n", email["subject"])

		// Simulate email sending
		time.Sleep(100 * time.Millisecond)

		fmt.Printf("[Worker] ✓ Email sent successfully!\n\n")
		return nil
	})

	// Start queue
	ctx := context.Background()
	if err := q.Start(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Queue started! Dispatching jobs...")

	// Dispatch some jobs
	for i := 1; i <= 5; i++ {
		job, err := q.Dispatch("send-email", map[string]interface{}{
			"to":      fmt.Sprintf("user%d@example.com", i),
			"subject": fmt.Sprintf("Welcome #%d", i),
			"body":    "Welcome to our service!",
		})
		if err != nil {
			log.Printf("Failed to dispatch job: %v", err)
			continue
		}
		fmt.Printf("[Dispatcher] Job %s dispatched\n", job.ID)
	}

	// Let jobs process
	fmt.Println("\nProcessing jobs...")
	time.Sleep(2 * time.Second)

	// Stop queue
	fmt.Println("\nStopping queue...")
	if err := q.Stop(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Queue stopped gracefully")
}
