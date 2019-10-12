package filter

// MapKey is a filter only pass when specified key found
type MapKey struct {
}

// Filter will only pass when key found
func (flt *MapKey) Filter(triggerParam, filterParam map[string]interface{}) (map[string]interface{}, bool) {
	var requiredKeys []string

	mapKey, ok := filterParam["map_key"]
	if ok {
		value, ok := mapKey.(string)
		if ok {
			requiredKeys = []string{value}
		} else {
			requiredKeys, _ = mapKey.([]string)
		}
	}

	for _, key := range requiredKeys {
		_, ok := triggerParam[key]
		if !ok {
			return nil, false
		}
	}

	return triggerParam, true
}
