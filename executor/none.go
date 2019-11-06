package executor

// None is a runner just do nothing
type None struct {
}

// Execute will do nothing
func (exe *None) Execute(param map[string]interface{}) map[string]interface{} {
	return nil
}

func newExecutorNone(param map[string]interface{}) interface{} {
	return &None{}
}
