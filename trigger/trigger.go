package trigger

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xianghuzhao/herald"
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
func CreateTrigger(name string) (herald.Trigger, error) {
	typeTrigger, ok := mapTrigger[name]
	if !ok {
		return nil, fmt.Errorf("Trigger \"%s\" not found", name)
	}
	tgrI := reflect.New(typeTrigger.Elem()).Interface()
	tgr, ok := tgrI.(herald.Trigger)
	if !ok {
		return nil, fmt.Errorf("\"%s\" is not a Trigger", name)
	}
	return tgr, nil
}
