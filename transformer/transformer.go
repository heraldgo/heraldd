package transformer

import (
	"fmt"
)

var transformers = map[string]func(map[string]interface{}) interface{}{
	"empty": newTransformerEmpty,
}

// CreateTransformer create a new transformer
func CreateTransformer(name string, param map[string]interface{}) (interface{}, error) {
	transformerCreator, ok := transformers[name]
	if !ok {
		return nil, fmt.Errorf(`Transformer "%s" not found`, name)
	}
	tfm := transformerCreator(param)
	return tfm, nil
}
