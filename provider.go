package dgqueue

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/donnigundala/dg-core/contracts/foundation"
)

// QueueServiceProvider implements the PluginProvider interface.
// This provides a simple, plug-and-play integration for applications.
//
// The provider expects the application to configure the driver.
// For automatic driver configuration, applications should use a wrapper provider.
//
// For advanced use cases requiring custom drivers or configuration,
// use the library functions (New, SetDriver) directly.
type QueueServiceProvider struct {
	// Config holds queue configuration
	// Auto-injected by dg-core if using config:"queue" tag
	Config Config `config:"queue"`

	// DriverFactory is an optional function to create the driver
	// If nil, the driver must be set manually after registration
	DriverFactory func(Config) (Driver, error)
}

// NewQueueServiceProvider creates a new queue service provider.
func NewQueueServiceProvider(driverFactory func(Config) (Driver, error)) *QueueServiceProvider {
	return &QueueServiceProvider{
		DriverFactory: driverFactory,
	}
}

// Name returns the name of the plugin.
func (p *QueueServiceProvider) Name() string {
	return Binding
}

// Version returns the version of the plugin.
func (p *QueueServiceProvider) Version() string {
	return Version
}

// Dependencies returns the list of dependencies.
func (p *QueueServiceProvider) Dependencies() []string {
	return []string{}
}

// Register registers the queue service provider.
func (p *QueueServiceProvider) Register(app foundation.Application) error {
	app.Singleton(Binding, func() (interface{}, error) {
		// Use provided config or default
		cfg := p.Config
		if cfg.Driver == "" {
			cfg = DefaultConfig()
		}

		// Try to resolve logger (optional)
		if cfg.Logger == nil {
			if loggerInstance, err := app.Make("logger"); err == nil {
				// Adapt dg-core logger to queue.Logger interface
				if adapted, ok := loggerInstance.(interface {
					Debug(msg string, args ...interface{})
					Info(msg string, args ...interface{})
					Warn(msg string, args ...interface{})
					Error(msg string, args ...interface{})
				}); ok {
					cfg.Logger = &loggerAdapter{logger: adapted}
				}
			}
		}

		// Create the manager
		manager := New(cfg)

		// Resolve driver
		var driver Driver
		if p.DriverFactory != nil {
			var err error
			driver, err = p.DriverFactory(cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to create queue driver from factory: %w", err)
			}
		} else {
			// Use global registry
			globalDriversMu.RLock()
			factory, ok := globalDrivers[cfg.Driver]
			globalDriversMu.RUnlock()

			if ok {
				var err error
				driver, err = factory(cfg)
				if err != nil {
					return nil, fmt.Errorf("failed to create queue driver %s: %w", cfg.Driver, err)
				}
			}
		}

		if driver != nil {
			manager.SetDriver(driver)
		} else if cfg.Driver != "" {
			return nil, fmt.Errorf("queue driver %s not found and no factory provided", cfg.Driver)
		}

		return manager, nil
	})

	return nil
}

// Boot boots the queue service provider.
func (p *QueueServiceProvider) Boot(app foundation.Application) error {
	return nil
}

// Shutdown gracefully stops the queue manager.
func (p *QueueServiceProvider) Shutdown(app foundation.Application) error {
	queueInstance, err := app.Make("queue")
	if err != nil {
		return nil // Queue not initialized
	}

	manager := queueInstance.(*Manager)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return manager.Stop(ctx)
}

// loggerAdapter adapts a generic logger to queue.Logger interface.
type loggerAdapter struct {
	logger interface {
		Debug(msg string, args ...interface{})
		Info(msg string, args ...interface{})
		Warn(msg string, args ...interface{})
		Error(msg string, args ...interface{})
	}
}

func (l *loggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *loggerAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *loggerAdapter) With(args ...interface{}) Logger {
	// Try to call With(args...) via reflection to support different return types
	v := reflect.ValueOf(l.logger)
	m := v.MethodByName("With")
	if m.IsValid() {
		valArgs := make([]reflect.Value, len(args))
		for i, arg := range args {
			valArgs[i] = reflect.ValueOf(arg)
		}
		results := m.Call(valArgs)
		if len(results) == 1 {
			if nextLogger, ok := results[0].Interface().(interface {
				Debug(msg string, args ...interface{})
				Info(msg string, args ...interface{})
				Warn(msg string, args ...interface{})
				Error(msg string, args ...interface{})
			}); ok {
				return &loggerAdapter{logger: nextLogger}
			}
		}
	}
	return l
}
