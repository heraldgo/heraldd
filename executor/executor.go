package executor

import (
	"reflect"
	"strings"
)

var executors = []interface{}{
	(*Print)(nil),
}

var mapExecutor map[string]reflect.Type

func init() {
	mapExecutor = make(map[string]reflect.Type)
	for _, method := range executors {
		methodName := strings.ToLower(reflect.TypeOf(method).Elem().Name())
		mapExecutor[methodName] = reflect.TypeOf(method)
	}
}

// CreateExecutor create a new executor
func CreateExecutor(name string) (interface{}, error) {
	typeExecutor, ok := mapExecutor[name]
	if !ok {
		return nil, nil
	}
	return reflect.New(typeExecutor.Elem()).Interface(), nil
}
