package executor

import (
	"github.com/xianghuzhao/heraldd/util"
)

// ExeClient is an executor which runs jobs locally
type ExeClient struct {
	util.BaseLogger
	Host    string
	WorkDir string
}

// Execute will run job locally
func (exe *ExeClient) Execute(param map[string]interface{}) map[string]interface{} {
	return nil
}

// SetParam will set param from a map
func (exe *ExeClient) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.Host, param, "host")
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
}
