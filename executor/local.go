package executor

import (
	"github.com/heraldgo/heraldd/util"
)

// Local is an executor which runs jobs locally
type Local struct {
	util.ExeGit
}

// Execute will execute job locally
func (exe *Local) Execute(param map[string]interface{}) map[string]interface{} {
	return exe.ExeGit.Execute(param)
}

// SetParam will set param from a map
func (exe *Local) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
}
