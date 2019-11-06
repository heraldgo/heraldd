package transformer

// Empty is a transformer get empty result
type Empty struct {
}

// Transform will return empty param
func (tfm *Empty) Transform(triggerParam map[string]interface{}) map[string]interface{} {
	return nil
}

func newTransformerEmpty(param map[string]interface{}) interface{} {
	return &Empty{}
}
