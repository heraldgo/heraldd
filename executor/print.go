package executor

import (
	"log"
)

// Print is a runner just print the param
type Print struct {
}

// Execute will print the param
func (exe *Print) Execute(param map[string]interface{}) map[string]interface{} {
	log.Printf("[Executor:Print] Execute with param:\n%#v\n", param)
	return nil
}
