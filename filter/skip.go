package filter

// Skip is a filter skip counters regularly
type Skip struct {
}

// Filter will skip certain numbers
func (flt *Skip) Filter(triggerParam, filterParam map[string]interface{}) (map[string]interface{}, bool) {
	skipNumber, ok := filterParam["skip_number"]
	if !ok {
		return triggerParam, true
	}
	skipNumberInt, ok := skipNumber.(int)
	if !ok || skipNumberInt <= 0 {
		return triggerParam, true
	}

	counter, ok := triggerParam["counter"]
	if !ok {
		return nil, false
	}
	counterInt, ok := counter.(int)
	if !ok || counterInt%(skipNumberInt+1) != 0 {
		return nil, false
	}

	return triggerParam, true
}

func newFilterSkip(map[string]interface{}) interface{} {
	return &Skip{}
}
