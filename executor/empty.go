package executor

// Empty is a runner just do nothing
type Empty struct {
}

// Execute will do nothing
func (exe *Empty) Execute(param map[string]interface{}) map[string]interface{} {
	return nil
}

func newExecutorEmpty(param map[string]interface{}) interface{} {
	return &Empty{}
}
