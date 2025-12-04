package queue

import (
	"context"
	"fmt"
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

// Name returns the name of the plugin.
func (p *QueueServiceProvider) Name() string {
	return "queue"
}

// Version returns the version of the plugin.
func (p *QueueServiceProvider) Version() string {
	return "1.4.0"
}

// Dependencies returns the list of dependencies.
func (p *QueueServiceProvider) Dependencies() []string {
	return []string{}
}

// Register registers the queue service provider.
func (p *QueueServiceProvider) Register(app foundation.Application) error {
	// Use provided config or default
	cfg := p.Config
	if cfg.Driver == "" {
		cfg = DefaultConfig()
	}

	// Try to resolve logger (optional)
	if cfg.Logger == nil {
		if loggerInstance, err := app.Make("logger"); err == nil {
			// Adapt dg-core logger to queue.Logger interface
			if logger, ok := loggerInstance.(interface {
				Info(msg string, keysAndValues ...interface{})
				Error(msg string, keysAndValues ...interface{})
				Warn(msg string, keysAndValues ...interface{})
			}); ok {
				cfg.Logger = &loggerAdapter{logger: logger}
			}
		}
	}

	// Register the queue manager as a singleton
	app.Singleton("queue", func() (interface{}, error) {
		manager := New(cfg)

		// If a driver factory is provided, use it
		if p.DriverFactory != nil {
			driver, err := p.DriverFactory(cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to create queue driver: %w", err)
			}
			manager.SetDriver(driver)
		}
		// Otherwise, driver must be set by the application

		return manager, nil
	})

	return nil
}

// Boot boots the queue service provider.
func (p *QueueServiceProvider) Boot(app foundation.Application) error {
	// Queue is started manually by the application
	// This allows the app to register workers before starting
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
		Info(msg string, keysAndValues ...interface{})
		Error(msg string, keysAndValues ...interface{})
		Warn(msg string, keysAndValues ...interface{})
	}
}

func (l *loggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *loggerAdapter) Error(msg string, err error, keysAndValues ...interface{}) {
	args := append([]interface{}{"error", err}, keysAndValues...)
	l.logger.Error(msg, args...)
}

func (l *loggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, keysAndValues...)
}
