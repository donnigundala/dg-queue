package main

import (
	"context"
	"fmt"
	"time"

	contractFoundation "github.com/donnigundala/dg-core/contracts/foundation"
	"github.com/donnigundala/dg-core/foundation"
	dgqueue "github.com/donnigundala/dg-queue"
)

// UserService demonstrates dependency injection
type UserService struct {
	*dgqueue.Injectable
}

func NewUserService(app contractFoundation.Application) *UserService {
	return &UserService{
		Injectable: dgqueue.NewInjectable(app),
	}
}

func (s *UserService) Welcome(ctx context.Context, email string) {
	fmt.Printf("Registering user %s...\n", email)

	// Use queue via Injectable
	job, err := s.Queue().Dispatch(ctx, "send-email", map[string]string{
		"email": email,
		"type":  "welcome",
	})

	if err != nil {
		fmt.Printf("Error queueing job: %v\n", err)
		return
	}

	fmt.Printf("Queued job %s\n", job.ID)
}

func main() {
	// 1. Setup Container
	app := foundation.New(".")

	// 2. Setup Config (simulated)
	cfg := dgqueue.DefaultConfig()
	cfg.Driver = "memory"

	// 3. Register Provider (Manual registration for example)
	// In a real app, this is done by the framework
	// 3. Register Provider
	provider := dgqueue.NewQueueServiceProvider(nil)
	provider.Config = cfg // Manually set config for this example
	app.Register(provider)

	// 4. Use Helper Functions
	q := dgqueue.MustResolve(app)
	q.Start()
	defer q.Stop(context.Background()) // ctx is optional for Stop in this example

	// Register a worker to see it working
	q.Worker("send-email", 1, func(ctx context.Context, job *dgqueue.Job) error {
		payload := job.Payload.(map[string]interface{})
		fmt.Printf("[Worker] Sending %s email to %s\n", payload["type"], payload["email"])
		return nil
	})

	// 5. Use Dependency Injection Service
	userService := NewUserService(app)
	userService.Welcome(context.Background(), "john@example.com")

	// Wait for worker to process
	time.Sleep(1 * time.Second)
}
