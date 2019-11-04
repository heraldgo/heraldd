package filter

import (
	"strings"

	"github.com/heraldgo/heraldd/util"
)

// MapKey is a filter only pass when specified key found
type MapKey struct {
}

// Filter will only pass when key found
func (flt *MapKey) Filter(triggerParam, filterParam map[string]interface{}) (map[string]interface{}, bool) {
	matchKey, err := util.GetStringParam(filterParam, "match_key")
	if err != nil {
		return triggerParam, true
	}

	var currentParam interface{}
	currentParam = triggerParam
	for _, frag := range strings.Split(matchKey, "/") {
		currentMap, ok := currentParam.(map[string]interface{})
		if !ok {
			return nil, false
		}
		currentParam, ok = currentMap[frag]
		if !ok {
			return nil, false
		}
	}
	foundValue, _ := currentParam.(string)

	matchValue, err := util.GetStringParam(filterParam, "match_value")
	if err == nil && foundValue != matchValue {
		return nil, false
	}

	return triggerParam, true
}

func newFilterMatchMap(map[string]interface{}) interface{} {
	return &MapKey{}
}
