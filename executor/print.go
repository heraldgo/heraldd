package executor

import (
	"github.com/xianghuzhao/herald"
)

// Print is a runner just print the param
type Print struct {
	log herald.Logger
}

// Execute will print the param
func (exe *Print) Execute(param map[string]interface{}) map[string]interface{} {
	exe.log.Infof("[Executor:Print] Execute with param:\n%#v\n", param)
	return nil
}

// SetLogger will set logger
func (exe *Print) SetLogger(logger herald.Logger) {
	exe.log = logger
}
