package executor

import (
	"github.com/xianghuzhao/heraldd/util"
)

// Local is an executor which runs jobs locally
type Local struct {
	util.ExeGit
	ExtraMap map[string]interface{}
}

// Execute will execute job locally
func (exe *Local) Execute(param map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range exe.ExtraMap {
		result[k] = v
	}

	for k, v := range exe.ExeGit.Execute(param) {
		result[k] = v
	}

	return result
}

// SetParam will set param from a map
func (exe *Local) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
	util.UpdateMapParam(&exe.ExtraMap, param, "extra_map")
}
