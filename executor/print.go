package executor

import (
	"encoding/json"

	"github.com/heraldgo/heraldd/util"
)

// Print is a runner just print the param
type Print struct {
	util.BaseLogger
}

// Execute will print the param
func (exe *Print) Execute(param map[string]interface{}) map[string]interface{} {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		exe.Errorf("Convert param argument failed: %s", err)
		return nil
	}
	exe.Infof("Execute with param: %s", paramJSON)
	return nil
}

func newExecutorPrint(param map[string]interface{}) interface{} {
	return &Print{}
}
