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
	var resultParam map[string]interface{}

	if len(exe.Keys) == 0 {
		resultParam = param
	} else {
		resultParam = make(map[string]interface{})
		for _, key := range exe.Keys {
			value, err := util.GetNestedMapValue(param, key)
			if err != nil {
				continue
			}
			resultParam[key] = value
		}
	}

	paramJSON, err := json.Marshal(resultParam)
	if err != nil {
		exe.Errorf("Convert param argument failed: %s", err)
		return nil
	}
	exe.Infof("Execute param: %s", paramJSON)
	return nil
}

func newExecutorPrint(param map[string]interface{}) interface{} {
	keys, _ := util.GetStringSliceParam(param, "key")
	return &Print{
		Keys: keys,
	}
}
