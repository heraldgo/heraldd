package selector

import (
	"encoding/json"

	"github.com/heraldgo/heraldd/util"
)

// External is a selector call a sub process to check the result
type External struct {
	util.BaseLogger
	Program string
}

// Select will call a sub process to check the exit code
func (slt *External) Select(triggerParam, selectorParam map[string]interface{}) bool {
	triggerParamJSON, err := json.Marshal(triggerParam)
	if err != nil {
		slt.Errorf("Generate trigger param argument failed: %s", err)
		return false
	}

	selectorParamJSON, err := json.Marshal(selectorParam)
	if err != nil {
		slt.Errorf("Generate selector param argument failed: %s", err)
		return false
	}

	exitCode, err := util.RunCmd([]string{slt.Program, string(triggerParamJSON), string(selectorParamJSON)}, "", false, nil, nil)
	if exitCode != 0 || err != nil {
		slt.Debugf("External selector does not pass: exit(%d), err(%s)", exitCode, err)
		return false
	}

	return true
}

func newSelectorExternal(param map[string]interface{}) interface{} {
	program, _ := util.GetStringParam(param, "program")

	return &External{
		Program: program,
	}
}
