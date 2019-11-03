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
func CreateTrigger(name string, param map[string]interface{}) (interface{}, error) {
	triggerCreator, ok := triggers[name]
	if !ok {
		return nil, fmt.Errorf(`Trigger "%s" not found`, name)
	}
	tgr := triggerCreator(param)
	return tgr, nil
}
