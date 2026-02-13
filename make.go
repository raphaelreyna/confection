package confection

import (
	"context"
	"fmt"
	"reflect"
)

// MakeCtx constructs an implementation of interface I from the given TypedConfig,
// using the provided context. Pass nil for c to use the global registry.
func MakeCtx[I Interface](ctx context.Context, c *Confection, tc TypedConfig) (I, error) {
	conf := getConfection(c)

	conf.mu.RLock()
	defer conf.mu.RUnlock()

	var iface I
	interfaceName := reflect.TypeFor[I]().String()
	apiObj, ok := conf.interfaces[interfaceName]
	if !ok {
		return iface, fmt.Errorf("line %d: interface %s not registered", tc.line, interfaceName)
	}

	newImplFunc, exists := apiObj.registeredTypes[tc.Type()]
	if !exists {
		return iface, fmt.Errorf("line %d: config type %q not registered for interface %s", tc.line, tc.Type(), interfaceName)
	}

	newImpl, err := newImplFunc(ctx, tc.TypedConfig)
	if err != nil {
		return iface, fmt.Errorf("line %d: %w", tc.line, err)
	}
	x, ok := newImpl.(I)
	if !ok {
		return iface, fmt.Errorf("line %d: factory for %q returned %T, which does not implement %s", tc.line, tc.Type(), newImpl, interfaceName)
	}

	return x, nil
}

// Make constructs an implementation of interface I from the given TypedConfig,
// using context.Background(). Pass nil for c to use the global registry.
func Make[I Interface](c *Confection, tc TypedConfig) (I, error) {
	ctx := context.Background()
	return MakeCtx[I](ctx, c, tc)
}
