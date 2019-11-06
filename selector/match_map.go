package selector

import (
	"github.com/heraldgo/heraldd/util"
)

// MapKey is a selector only pass when specified key found
type MapKey struct {
}

// Select will only pass when key found
func (slt *MapKey) Select(triggerParam, selectorParam map[string]interface{}) bool {
	matchKey, err := util.GetStringParam(selectorParam, "match_key")
	if err != nil {
		return true
	}

	foundValue, err := util.GetNestedMapValue(triggerParam, matchKey)
	if err != nil {
		return false
	}

	matchValue, ok := selectorParam["match_value"]
	if !ok {
		return true
	}

	return foundValue == matchValue
}

func newSelectorMatchMap(map[string]interface{}) interface{} {
	return &MapKey{}
}
