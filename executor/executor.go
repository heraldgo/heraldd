package executor

import (
	"fmt"
	"reflect"

	"github.com/xianghuzhao/herald"

	"github.com/xianghuzhao/heraldd/util"
)

var executors = []interface{}{
	(*Print)(nil),
	(*Local)(nil),
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
func CreateExecutor(name string) (herald.Executor, error) {
	typeExecutor, ok := mapExecutor[name]
	if !ok {
		return nil, fmt.Errorf("Executor \"%s\" not found", name)
	}
	exeI := reflect.New(typeExecutor.Elem()).Interface()
	exe, ok := exeI.(herald.Executor)
	if !ok {
		return nil, fmt.Errorf("\"%s\" is not an Executor", name)
	}
	return exe, nil
}
