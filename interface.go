package confection

import (
	"reflect"
)

// Interface is the marker interface that all confection-managed interfaces must embed.
// Implementation structs must embed the target interface as a field so that
// confection can auto-discover which interfaces a factory satisfies.
type Interface interface {
	_isConfectionInterface()
}

var _interfaceType = reflect.TypeOf((*Interface)(nil)).Elem()
