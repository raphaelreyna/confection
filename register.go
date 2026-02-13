package confection

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// RegisterInterface registers an interface type with the given Confection registry.
// Pass nil to use the global registry.
// Panics if the interface is already registered.
func RegisterInterface[I Interface](c *Confection) {
	conf := getConfection(c)

	name := reflect.TypeFor[I]().String()

	conf.mu.Lock()
	defer conf.mu.Unlock()

	if _, ok := conf.interfaces[name]; ok {
		panic(fmt.Sprintf("unable to register Interface %q: Interface already registered", name))
	}

	conf.interfaces[name] = &_interface{}
}

// Factory is a function that creates an Implementation from a Configuration.
type Factory[Configuration any, Implementation any] func(context.Context, Configuration) (Implementation, error)

// RegisterFactory registers a factory function that creates an Implementation
// from a Configuration, binding it to the given @type name.
// The Implementation must be a pointer to a struct that embeds one or more
// registered confection Interface types.
// Pass nil for c to use the global registry.
// Panics on invalid types, unregistered interfaces, or duplicate registrations.
func RegisterFactory[Configuration any, Implementation any](c *Confection, typeName string, factory Factory[Configuration, Implementation]) {
	conf := getConfection(c)

	t := reflect.TypeFor[Implementation]()
	if t.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("unable to register factory: type %s is not a pointer to a struct", t))
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("unable to register factory: type %s is not a struct", t))
	}

	// find all Interfaces that the output type implements
	usingPositiveTagging := false
	interfaceNamesAndTags := make(map[string]string, 0)
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

		tag := field.Tag.Get("confection")
		if strings.Contains(tag, "implement") {
			usingPositiveTagging = true
			tag = "implement"
		}
		if tag != "implement" && tag != "-" && tag != "" {
			panic(fmt.Sprintf("unable to register factory with config name %q for Interface %q: invalid tag %q", typeName, ifaceType.String(), tag))
		}
		interfaceNamesAndTags[ifaceType.String()] = tag
	}

	interfaceNames := make([]string, 0)
	if usingPositiveTagging {
		for interfaceName, tag := range interfaceNamesAndTags {
			if tag == "implement" {
				interfaceNames = append(interfaceNames, interfaceName)
			}
		}
	} else {
		for interfaceName, tag := range interfaceNamesAndTags {
			if tag != "-" {
				interfaceNames = append(interfaceNames, interfaceName)
			}
		}
	}

	conf.mu.Lock()
	defer conf.mu.Unlock()

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
