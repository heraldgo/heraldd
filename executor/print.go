package executor

import (
	"github.com/xianghuzhao/heraldd/util"
)

// Print is a runner just print the param
type Print struct {
	util.BaseLogger
}

// Execute will print the param
func (exe *Print) Execute(param map[string]interface{}) map[string]interface{} {
	exe.Infof("[Executor(Print)] Execute with param: %#v", param)
	return nil
}
