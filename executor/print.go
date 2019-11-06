package executor

import (
	"encoding/json"

	"github.com/heraldgo/heraldd/util"
)

// Print is a runner just print the param
type Print struct {
	util.BaseLogger
	Keys []string
}

// Execute will print the param
func (exe *Print) Execute(param map[string]interface{}) map[string]interface{} {
	resultParam := make(map[string]interface{})
	for k, v := range param {
		for _, key := range exe.Keys {
			if k == key {
				resultParam[k] = v
				break
			}
		}
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		exe.Errorf("Convert param argument failed: %s", err)
		return nil
	}
	exe.Infof("Execute with param: %s", paramJSON)
	return nil
}

func newExecutorPrint(param map[string]interface{}) interface{} {
	keys, _ := util.GetStringSliceParam(param, "key")
	return &Print{
		Keys: keys,
	}
}
