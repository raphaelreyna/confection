package confection

import (
	"context"
	"sync"

	"gopkg.in/yaml.v3"
)

type _interface struct {
	registeredTypes map[string]func(context.Context, *yaml.Node) (any, error)
}

// Confection is a typed configuration registry that maps interface types
// to factory functions, enabling config-driven polymorphism.
type Confection struct {
	mu         sync.RWMutex
	interfaces map[string]*_interface
}

func (c *Confection) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := ""
	for k, v := range c.interfaces {
		if len(v.registeredTypes) == 0 {
			s += "Interface " + k + " has no registered types\n"
			continue
		}
		s += "Interface: " + k + "\n"
		for k := range v.registeredTypes {
			s += "  Config: " + k + "\n"
		}
	}
	return s
}

// NewConfection creates a new, empty Confection registry.
func NewConfection() *Confection {
	c := Confection{
		interfaces: make(map[string]*_interface, 0),
	}

	return &c
}

// Global is the package-level default registry.
// It is lazily initialized on first use when nil is passed to any API function.
var Global *Confection

var globalOnce sync.Once

func getGlobal() *Confection {
	globalOnce.Do(func() {
		if Global == nil {
			Global = NewConfection()
		}
	})
	return Global
}

func getConfection(c *Confection) *Confection {
	if c == nil {
		return getGlobal()
	}
	return c
}
