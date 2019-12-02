package selector

// All is a selector pass all the triggers
type All struct {
}

// Select will always pass
func (slt *All) Select(triggerParam, jobParam map[string]interface{}) bool {
	return true
}

func newSelectorAll(map[string]interface{}) interface{} {
	return &All{}
}
