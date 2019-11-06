package filter

// Skip is a filter skip counters regularly
type Skip struct {
}

// Filter will skip certain numbers
func (flt *Skip) Filter(triggerParam, filterParam map[string]interface{}) bool {
	skipNumber, ok := filterParam["skip_number"]
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

func newFilterSkip(map[string]interface{}) interface{} {
	return &Skip{}
}
