package executor

import (
	"fmt"
)

var executors = map[string]func(map[string]interface{}) interface{}{
	"none":        newExecutorNone,
	"print":       newExecutorPrint,
	"local":       newExecutorLocal,
	"http_remote": newExecutorHTTPRemote,
}

// CreateExecutor create a new executor
func CreateExecutor(name string, param map[string]interface{}) (interface{}, error) {
	executorCreator, ok := executors[name]
	if !ok {
		return nil, fmt.Errorf(`Executor "%s" not found`, name)
	}
	exe := executorCreator(param)
	return exe, nil
}
