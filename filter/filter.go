package filter

import (
	"fmt"
)

var filters = map[string]func(map[string]interface{}) interface{}{
	"skip":    newFilterSkip,
	"map_key": newFilterMapKey,
}

// CreateFilter create a new filter
func CreateFilter(name string, param map[string]interface{}) (interface{}, error) {
	filterCreator, ok := filters[name]
	if !ok {
		return nil, fmt.Errorf(`Filter "%s" not found`, name)
	}
	flt := filterCreator(param)
	return flt, nil
}
