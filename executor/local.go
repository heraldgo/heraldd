package executor

import (
	"fmt"

	"github.com/heraldgo/heraldd/util"
)

// Local is an executor which runs jobs locally
type Local struct {
	util.ExeGit
}

// Execute will execute job locally
func (exe *Local) Execute(param map[string]interface{}) (map[string]interface{}, error) {
	result, err := exe.ExeGit.Execute(param)
	if err != nil {
		return result, err
	}

	exitCode, _ := util.GetIntParam(result, "exit_code")

	if exitCode != 0 {
		return result, fmt.Errorf("Command failed with code %d", exitCode)
	}

	return result, nil
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
