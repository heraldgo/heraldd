package selector

import (
	"github.com/heraldgo/heraldd/util"
)

// MatchMap is a selector only pass when specified key found
type MatchMap struct {
}

// Select will only pass when key found
func (slt *MatchMap) Select(triggerParam, jobParam map[string]interface{}) bool {
	matchKey, err := util.GetStringParam(jobParam, "match_key")
	if err != nil {
		return false
	}

	foundValue, err := util.GetNestedMapValue(triggerParam, matchKey)
	if err != nil {
		return false
	}

	matchValue, ok := jobParam["match_value"]
	if !ok {
		return true
	}

	return foundValue == matchValue
}

func newSelectorMatchMap(map[string]interface{}) interface{} {
	return &MatchMap{}
}
