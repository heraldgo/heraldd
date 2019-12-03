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
func CreateExecutor(typeName string, param map[string]interface{}) (interface{}, error) {
	executorCreator, ok := executors[typeName]
	if !ok {
		return nil, fmt.Errorf(`Executor "%s" not found`, typeName)
	}
	exe := executorCreator(param)
	return exe, nil
}
