package confection

import (
	"context"
	"fmt"
	"reflect"
)

func MakeCtx[I Interface](ctx context.Context, c *Confection, tc TypedConfig) (I, error) {
	conf := getConfection(c)

	var iface I
	interfaceName := reflect.TypeFor[I]().String()
	apiObj, ok := conf.interfaces[interfaceName]
	if !ok {
		return iface, fmt.Errorf("API %s not found", interfaceName)
	}

	newImplFunc, exists := apiObj.registeredTypes[tc.Type()]
	if !exists {
		return iface, fmt.Errorf("config type %s not registered for API %s", tc.Type(), interfaceName)
	}

	newImpl, err := newImplFunc(ctx, tc.TypedConfig)
	if err != nil {
		return iface, err
	}
	x, ok := newImpl.(I)
	if !ok {
		return iface, fmt.Errorf("unable to cast %T to %T", newImpl, iface)
	}

	return x, nil
}

func Make[I Interface](c *Confection, tc TypedConfig) (I, error) {
	ctx := context.Background()
	return MakeCtx[I](ctx, c, tc)
}
