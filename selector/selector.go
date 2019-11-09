package selector

import (
	"fmt"
)

var selectors = map[string]func(map[string]interface{}) interface{}{
	"all":       newSelectorAll,
	"skip":      newSelectorSkip,
	"match_map": newSelectorMatchMap,
	"external":  newSelectorExternal,
}

// CreateSelector create a new selector
func CreateSelector(name string, param map[string]interface{}) (interface{}, error) {
	selectorCreator, ok := selectors[name]
	if !ok {
		return nil, fmt.Errorf(`Selector "%s" not found`, name)
	}
	slt := selectorCreator(param)
	return slt, nil
}
