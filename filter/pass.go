package filter

// Pass is a filter pass the trigger param directly
type Pass struct {
}

// Filter will pass the trigger param
func (flt *Pass) Filter(triggerParam, filterParam map[string]interface{}) (map[string]interface{}, bool) {
	return triggerParam, true
}

func newFilterPass(map[string]interface{}) interface{} {
	return &Pass{}
}
