package confection

import (
	"reflect"
)

type Interface interface {
	_isConfectionInterface()
}

var _interfaceType = reflect.TypeOf((*Interface)(nil)).Elem()
