package selector

// Skip is a selector skip counters regularly
type Skip struct {
}

// Select will skip certain numbers
func (slt *Skip) Select(triggerParam, selectParam map[string]interface{}) bool {
	skipNumber, ok := selectParam["skip_number"]
	if !ok {
		return true
	}
	skipNumberInt, ok := skipNumber.(int)
	if !ok || skipNumberInt <= 0 {
		return true
	}

	counter, ok := triggerParam["counter"]
	if !ok {
		return false
	}
	counterInt, ok := counter.(int)
	if !ok || counterInt%(skipNumberInt+1) != 0 {
		return false
	}

	return true
}

func newSelectorSkip(map[string]interface{}) interface{} {
	return &Skip{}
}
