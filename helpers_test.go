package dgqueue

import (
	"testing"

	"github.com/donnigundala/dg-core/foundation"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	app := foundation.New(".")
	cfg := DefaultConfig()
	manager := New(cfg)

	app.Instance("queue", manager)

	// Test Resolve
	q, err := Resolve(app)
	assert.NoError(t, err)
	assert.NotNil(t, q)
	assert.Equal(t, manager, q)
}

func TestResolve_Error(t *testing.T) {
	app := foundation.New(".")

	// Test Resolve without registration
	q, err := Resolve(app)
	assert.Error(t, err)
	assert.Nil(t, q)
}

func TestMustResolve(t *testing.T) {
	app := foundation.New(".")
	cfg := DefaultConfig()
	manager := New(cfg)

	app.Instance("queue", manager)

	// Test MustResolve
	assert.NotPanics(t, func() {
		q := MustResolve(app)
		assert.Equal(t, manager, q)
	})
}

func TestMustResolve_Panic(t *testing.T) {
	app := foundation.New(".")

	// Test MustResolve panics without registration
	assert.Panics(t, func() {
		MustResolve(app)
	})
}

func TestInjectable(t *testing.T) {
	app := foundation.New(".")
	cfg := DefaultConfig()
	manager := New(cfg)

	app.Instance("queue", manager)

	inject := NewInjectable(app)

	assert.NotPanics(t, func() {
		q := inject.Queue()
		assert.Equal(t, manager, q)
	})
}

func TestInjectable_Panic(t *testing.T) {
	app := foundation.New(".")

	inject := NewInjectable(app)

	// Test Queue() panics without registration
	assert.Panics(t, func() {
		inject.Queue()
	})
}
