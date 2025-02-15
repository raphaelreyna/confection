package confection

import (
	"context"

	"gopkg.in/yaml.v3"
)

type _interface struct {
	_interface      Interface
	registeredTypes map[string]func(context.Context, *yaml.Node) (any, error)
}

type Confection struct {
	interfaces map[string]*_interface
}

func (c *Confection) String() string {
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

func NewConfection() *Confection {
	c := Confection{
		interfaces: make(map[string]*_interface, 0),
	}

	return &c
}

var Global *Confection

func getGlobal() *Confection {
	if Global == nil {
		Global = NewConfection()
	}
	return Global
}

func getConfection(c *Confection) *Confection {
	if c == nil {
		return getGlobal()
	}
	return c
}
