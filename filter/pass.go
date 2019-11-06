package filter

// Pass is a filter pass the trigger param directly
type Pass struct {
}

// Filter will always pass the check
func (flt *Pass) Filter(triggerParam, filterParam map[string]interface{}) bool {
	return true
}

func newFilterPass(map[string]interface{}) interface{} {
	return &Pass{}
}
