package filter

import (
	"fmt"
	"reflect"

	"github.com/xianghuzhao/heraldd/util"
)

var filters = []interface{}{
	(*Skip)(nil),
	(*MapKey)(nil),
}

var mapFilter map[string]reflect.Type

func init() {
	mapFilter = make(map[string]reflect.Type)
	for _, method := range filters {
		methodName := util.CamelToSnake(reflect.TypeOf(method).Elem().Name())
		mapFilter[methodName] = reflect.TypeOf(method)
	}
}

// CreateFilter create a new filter
func CreateFilter(name string) (interface{}, error) {
	typeFilter, ok := mapFilter[name]
	if !ok {
		return nil, fmt.Errorf("Filter \"%s\" not found", name)
	}
	flt := reflect.New(typeFilter.Elem()).Interface()
	return flt, nil
}
