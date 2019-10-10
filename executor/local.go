package executor

import (
	"log"
	"os"
	"os/exec"
)

// Local is a runner run jobs locally
type Local struct {
	WorkDir string
}

// Execute will run the local job
func (exe *Local) Execute(param map[string]interface{}) {
	err := os.MkdirAll(exe.WorkDir, 0755)
	if err != nil {
		log.Println(err)
	}

	jobParam, ok := param["job_param"]
	if ok {
		jobParamMap, ok := jobParam.(map[interface{}]interface{})
		if ok {
			scriptRepo, _ := jobParamMap["script_repo"].(string)
			if scriptRepo != "" {
				cmd := exec.Command("git", "clone", scriptRepo)
				cmd.Dir = exe.WorkDir
				err = cmd.Run()
				if err != nil {
					log.Printf("Run command error: %s\n", err)
				}
			}

			//routerArgument
		}
	}

	triggerParam, ok := param["trigger_param"]
	if ok {
		log.Println(triggerParam)
	}
}
