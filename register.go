package confection

import (
	"context"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

func RegisterInterface[I Interface](c *Confection) {
	conf := getConfection(c)

	var i I

	name := reflect.TypeFor[I]().String()

	if _, ok := conf.interfaces[name]; ok {
		panic(fmt.Sprintf("unable to register Interface %q: Interface already registered", name))
	}

	conf.interfaces[name] = &_interface{
		_interface: i,
	}
}

type Factory[Configuration any, Implementation any] func(context.Context, Configuration) (Implementation, error)

func RegisterFactory[Configuration any, Implementation any](c *Confection, typeName string, factory Factory[Configuration, Implementation]) {
	conf := getConfection(c)

	t := reflect.TypeFor[Implementation]()
	if t.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("unable register factory for Interface %q: type %T is not a pointer to a struct", t.Kind(), t))
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("unable register factory for Interface %q: type %T is not a struct", t.Kind(), t))
	}

	// find all Interfaces that the output type implements
	interfaceNames := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() != reflect.Interface {
			continue
		}

		// we found an interface, check if it is an Interface
		ifaceType := field.Type
		if ifaceType.Kind() != reflect.Interface {
			continue
		}
		if !ifaceType.Implements(_interfaceType) {
			continue
		}

		interfaceNames = append(interfaceNames, ifaceType.String())
	}

	// register the factory for each Interface under the config type name
	for _, interfaceName := range interfaceNames {
		iface, ok := conf.interfaces[interfaceName]
		if !ok {
			panic(fmt.Sprintf("unable to register factory with config name %q for Interface %q: Interface not found", typeName, interfaceName))
		}
		if iface.registeredTypes == nil {
			iface.registeredTypes = make(map[string]func(context.Context, *yaml.Node) (any, error))
		}

		_, exists := iface.registeredTypes[typeName]
		if exists {
			panic(fmt.Sprintf("unable to register factory with config name %q for Interface %q: configuration type double registration", typeName, interfaceName))
		}

		iface.registeredTypes[typeName] = func(ctx context.Context, node *yaml.Node) (any, error) {
			var config Configuration
			if err := node.Decode(&config); err != nil {
				return nil, err
			}
			return factory(ctx, config)
		}
	}
}
