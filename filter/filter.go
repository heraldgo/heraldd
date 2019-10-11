package filter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xianghuzhao/herald"
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
func CreateFilter(name string) (herald.Filter, error) {
	typeFilter, ok := mapFilter[name]
	if !ok {
		return nil, fmt.Errorf("Filter \"%s\" not found", name)
	}
	fltI := reflect.New(typeFilter.Elem()).Interface()
	flt, ok := fltI.(herald.Filter)
	if !ok {
		return nil, fmt.Errorf("\"%s\" is not a Filter", name)
	}
	return flt, nil
}
