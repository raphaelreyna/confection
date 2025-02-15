package confection

import (
	"reflect"
)

type Interface interface {
	_isConfectionAPI()
}

var _interfaceType = reflect.TypeOf((*Interface)(nil)).Elem()
