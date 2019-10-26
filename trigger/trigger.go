package trigger

import (
	"fmt"
	"reflect"

	"github.com/heraldgo/heraldd/util"
)

var triggers = []interface{}{
	(*Tick)(nil),
	(*Cron)(nil),
	(*HTTP)(nil),
}

var mapTrigger map[string]reflect.Type

func init() {
	mapTrigger = make(map[string]reflect.Type)
	for _, method := range triggers {
		methodName := util.CamelToSnake(reflect.TypeOf(method).Elem().Name())
		mapTrigger[methodName] = reflect.TypeOf(method)
	}
}

// CreateTrigger create a new trigger
func CreateTrigger(name string) (interface{}, error) {
	typeTrigger, ok := mapTrigger[name]
	if !ok {
		return nil, fmt.Errorf("Trigger \"%s\" not found", name)
	}
	tgr := reflect.New(typeTrigger.Elem()).Interface()
	return tgr, nil
}
