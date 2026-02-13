package dynamic

import (
	"fmt"
	"io"
	"sync"
)

// SourceFactory creates an io.ReadCloser from a YAML string value.
type SourceFactory func(value string) (io.ReadCloser, error)

// Registry holds registered data source types for use with DataSource.
type Registry struct {
	mu      sync.RWMutex
	sources map[string]SourceFactory
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]SourceFactory),
	}
}

// newDefaultRegistry creates a registry pre-loaded with the built-in source types.
func newDefaultRegistry() *Registry {
	r := NewRegistry()
	r.sources["file"] = fileSource
	r.sources["env"] = envSource
	r.sources["string"] = stringSource
	r.sources["bytes"] = bytesSource
	return r
}

// Global is the package-level default registry, pre-loaded with built-in sources
// (file, env, string, bytes).
var Global *Registry

var globalOnce sync.Once

func getGlobal() *Registry {
	globalOnce.Do(func() {
		if Global == nil {
			Global = newDefaultRegistry()
		}
	})
	return Global
}

func getRegistry(r *Registry) *Registry {
	if r == nil {
		return getGlobal()
	}
	return r
}

// RegisterSource registers a SourceFactory under the given name.
// Pass nil for r to use the global registry.
// Panics if the name is already registered.
func RegisterSource(r *Registry, name string, factory SourceFactory) {
	reg := getRegistry(r)
	reg.mu.Lock()
	defer reg.mu.Unlock()
	if _, exists := reg.sources[name]; exists {
		panic(fmt.Sprintf("data source type %q already registered", name))
	}
	reg.sources[name] = factory
}

func (reg *Registry) lookup(name string) (SourceFactory, bool) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	f, ok := reg.sources[name]
	return f, ok
}
