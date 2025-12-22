package dgqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueueServiceProvider_Name(t *testing.T) {
	provider := &QueueServiceProvider{}
	assert.Equal(t, "queue", provider.Name())
}

func TestQueueServiceProvider_Version(t *testing.T) {
	provider := &QueueServiceProvider{}
	assert.Equal(t, "1.6.0", provider.Version())
}

func TestQueueServiceProvider_Dependencies(t *testing.T) {
	provider := &QueueServiceProvider{}
	deps := provider.Dependencies()

	assert.NotNil(t, deps)
	assert.Empty(t, deps, "dg-queue should have no required dependencies")
}

func TestQueueServiceProvider_ConfigDefaults(t *testing.T) {
	provider := &QueueServiceProvider{}
	// Config should use zero values initially
	assert.Equal(t, "", provider.Config.Driver)
}

func TestQueueServiceProvider_CustomConfig(t *testing.T) {
	customConfig := Config{
		Driver:  "memory",
		Workers: 10,
	}

	provider := &QueueServiceProvider{
		Config: customConfig,
	}

	assert.Equal(t, "memory", provider.Config.Driver)
	assert.Equal(t, 10, provider.Config.Workers)
}
