package trigger

import (
	"fmt"
)

var triggers = map[string]func(map[string]interface{}) interface{}{
	"tick": newTriggerTick,
	"cron": newTriggerCron,
	"http": newTriggerHTTP,
}

// CreateTrigger create a new trigger
func CreateTrigger(typeName string, param map[string]interface{}) (interface{}, error) {
	triggerCreator, ok := triggers[typeName]
	if !ok {
		return nil, fmt.Errorf(`Trigger "%s" not found`, typeName)
	}
	tgr := triggerCreator(param)
	return tgr, nil
}
