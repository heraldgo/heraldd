package executor

import (
	"fmt"
	"reflect"

	"github.com/heraldgo/heraldd/util"
)

var executors = []interface{}{
	(*Print)(nil),
	(*Local)(nil),
	(*ExeClient)(nil),
}

var mapExecutor map[string]reflect.Type

func init() {
	mapExecutor = make(map[string]reflect.Type)
	for _, method := range executors {
		methodName := util.CamelToSnake(reflect.TypeOf(method).Elem().Name())
		mapExecutor[methodName] = reflect.TypeOf(method)
	}
}

// CreateExecutor create a new executor
func CreateExecutor(name string) (interface{}, error) {
	typeExecutor, ok := mapExecutor[name]
	if !ok {
		return nil, fmt.Errorf("Executor \"%s\" not found", name)
	}
	exe := reflect.New(typeExecutor.Elem()).Interface()
	return exe, nil
}
