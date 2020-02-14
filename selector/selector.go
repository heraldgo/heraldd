package selector

import (
	"fmt"
)

var selectors = map[string]func(map[string]interface{}) interface{}{
	"all":        newSelectorAll,
	"skip":       newSelectorSkip,
	"match_map":  newSelectorMatchMap,
	"except_map": newSelectorExceptMap,
	"external":   newSelectorExternal,
}

// CreateSelector create a new selector
func CreateSelector(typeName string, param map[string]interface{}) (interface{}, error) {
	selectorCreator, ok := selectors[typeName]
	if !ok {
		return nil, fmt.Errorf(`Selector "%s" not found`, typeName)
	}
	slt := selectorCreator(param)
	return slt, nil
}
