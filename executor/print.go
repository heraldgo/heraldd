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
func (exe *Print) Execute(param map[string]interface{}) (map[string]interface{}, error) {
	jobParam, _ := util.GetMapParam(param, "job_param")
	printKeys, _ := util.GetStringSliceParam(jobParam, "print_key")

	var resultParam map[string]interface{}

	if len(printKeys) == 0 {
		resultParam = param
	} else {
		resultParam = make(map[string]interface{})
		for _, key := range printKeys {
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
		return nil, nil
	}
	exe.Infof("Execute param: %s", paramJSON)
	return nil, nil
}

func newExecutorPrint(param map[string]interface{}) interface{} {
	return &Print{}
}
