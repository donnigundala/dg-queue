package dgqueue

import (
	"fmt"

	"github.com/donnigundala/dg-core/contracts/foundation"
)

// Resolve resolves the main queue manager from the application container.
func Resolve(app foundation.Application) (Queue, error) {
	instance, err := app.Make("queue")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve queue: %w", err)
	}

	queue, ok := instance.(Queue)
	if !ok {
		return nil, fmt.Errorf("resolved instance is not a Queue")
	}

	return queue, nil
}

// MustResolve resolves the queue manager or panics.
func MustResolve(app foundation.Application) Queue {
	queue, err := Resolve(app)
	if err != nil {
		panic(err)
	}
	return queue
}

// Injectable provides a convenient way to inject queue dependencies.
// Include this struct in your services to easily access the queue.
type Injectable struct {
	app foundation.Application
}

// NewInjectable creates a new Injectable instance.
func NewInjectable(app foundation.Application) *Injectable {
	return &Injectable{app: app}
}

// Queue returns the main queue manager.
// Panics if queue cannot be resolved.
func (i *Injectable) Queue() Queue {
	return MustResolve(i.app)
}
