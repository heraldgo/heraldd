package executor

import (
	"github.com/heraldgo/heraldd/util"
)

// Print is a runner just print the param
type Print struct {
	util.BaseLogger
}

// Execute will print the param
func (exe *Print) Execute(param map[string]interface{}) map[string]interface{} {
	exe.Infof("Execute with param: %#v", param)
	return nil
}

func newExecutorPrint(param map[string]interface{}) interface{} {
	return &Print{}
}
