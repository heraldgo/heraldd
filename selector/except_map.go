package selector

import (
	"github.com/heraldgo/heraldd/util"
)

// ExceptMap is a selector only pass when specified key not matched
type ExceptMap struct {
}

// Select will only pass when key not matched
func (slt *ExceptMap) Select(triggerParam, selectParam map[string]interface{}) bool {
	exceptKey, err := util.GetStringParam(selectParam, "except_key")
	if err != nil {
		return false
	}

	foundValue, err := util.GetNestedMapValue(triggerParam, exceptKey)
	if err != nil {
		return true
	}

	exceptValue, ok := selectParam["except_value"]
	if !ok {
		return false
	}

	return foundValue != exceptValue
}

func newSelectorExceptMap(map[string]interface{}) interface{} {
	return &ExceptMap{}
}
