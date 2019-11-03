package filter

import (
	"github.com/heraldgo/heraldd/util"
)

// MapKey is a filter only pass when specified key found
type MapKey struct {
}

// Filter will only pass when key found
func (flt *MapKey) Filter(triggerParam, filterParam map[string]interface{}) (map[string]interface{}, bool) {
	requiredKeys, _ := util.GetStringSliceParam(filterParam, "map_key")

	for _, key := range requiredKeys {
		_, ok := triggerParam[key]
		if !ok {
			return nil, false
		}
	}

	return triggerParam, true
}

func newFilterMapKey(map[string]interface{}) interface{} {
	return &MapKey{}
}
