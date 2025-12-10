package main

import (
	"fmt"
	"time"

	contractFoundation "github.com/donnigundala/dg-core/contracts/foundation"
	"github.com/donnigundala/dg-core/foundation"
	queue "github.com/donnigundala/dg-queue"
)

// UserService demonstrates dependency injection
type UserService struct {
	*queue.Injectable
}

func NewUserService(app contractFoundation.Application) *UserService {
	return &UserService{
		Injectable: queue.NewInjectable(app),
	}
}

func (s *UserService) Welcome(email string) {
	fmt.Printf("Registering user %s...\n", email)

	// Use queue via Injectable
	job, err := s.Queue().Dispatch("send-email", map[string]string{
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
	cfg := queue.DefaultConfig()
	cfg.Driver = "memory"

	// 3. Register Provider (Manual registration for example)
	// In a real app, this is done by the framework
	provider := &queue.QueueServiceProvider{
		Config: cfg,
	}
	if err := provider.Register(app); err != nil {
		panic(err)
	}

	// 4. Use Helper Functions
	q := queue.MustResolve(app)
	q.Start()
	defer q.Stop(nil) // ctx is optional for Stop in this example

	// Register a worker to see it working
	q.Worker("send-email", 1, func(job *queue.Job) error {
		payload := job.Payload.(map[string]string)
		fmt.Printf("[Worker] Sending %s email to %s\n", payload["type"], payload["email"])
		return nil
	})

	// 5. Use Dependency Injection Service
	userService := NewUserService(app)
	userService.Welcome("john@example.com")

	// Wait for worker to process
	time.Sleep(1 * time.Second)
}
