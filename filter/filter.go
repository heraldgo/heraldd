package filter

import (
	"reflect"
	"strings"
)

var filters = []interface{}{
	(*Skip)(nil),
}

var mapFilter map[string]reflect.Type

func init() {
	mapFilter = make(map[string]reflect.Type)
	for _, method := range filters {
		methodName := strings.ToLower(reflect.TypeOf(method).Elem().Name())
		mapFilter[methodName] = reflect.TypeOf(method)
	}
}

// CreateFilter create a new filter
func CreateFilter(name string) (interface{}, error) {
	typeFilter, ok := mapFilter[name]
	if !ok {
		return nil, nil
	}
	return reflect.New(typeFilter.Elem()).Interface(), nil
}
