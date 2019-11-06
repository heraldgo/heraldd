package filter

import (
	"github.com/heraldgo/heraldd/util"
)

// MapKey is a filter only pass when specified key found
type MapKey struct {
}

// Filter will only pass when key found
func (flt *MapKey) Filter(triggerParam, filterParam map[string]interface{}) bool {
	matchKey, err := util.GetStringParam(filterParam, "match_key")
	if err != nil {
		return true
	}

	foundValue, err := util.GetNestedMapValue(triggerParam, matchKey)
	if err != nil {
		return false
	}

	matchValue, ok := filterParam["match_value"]
	if !ok {
		return true
	}

	return foundValue == matchValue
}

func newFilterMatchMap(map[string]interface{}) interface{} {
	return &MapKey{}
}
