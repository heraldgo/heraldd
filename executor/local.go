package executor

import (
	"github.com/xianghuzhao/heraldd/util"
)

// Local is an executor which runs jobs locally
type Local struct {
	util.ExeGit
}

// SetParam will set param from a map
func (exe *Local) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
}
