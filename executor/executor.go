package executor

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xianghuzhao/herald"
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
