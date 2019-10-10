package trigger

import (
	"reflect"
	"strings"
)

var triggers = []interface{}{
	(*Tick)(nil),
	(*Cron)(nil),
}

var mapTrigger map[string]reflect.Type

func init() {
	mapTrigger = make(map[string]reflect.Type)
	for _, method := range triggers {
		methodName := strings.ToLower(reflect.TypeOf(method).Elem().Name())
		mapTrigger[methodName] = reflect.TypeOf(method)
	}
}

// CreateTrigger create a new trigger
func CreateTrigger(name string) (interface{}, error) {
	typeTrigger, ok := mapTrigger[name]
	if !ok {
		return nil, nil
	}
	return reflect.New(typeTrigger.Elem()).Interface(), nil
}
