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

func newExecutorLocal(param map[string]interface{}) interface{} {
	workDir, _ := util.GetStringParam(param, "work_dir")

	exe := &Local{
		ExeGit: util.ExeGit{
			WorkDir: workDir,
		},
	}
	return exe
}
