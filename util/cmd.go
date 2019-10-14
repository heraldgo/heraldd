package util

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunCmd will open the sub process
func RunCmd(args []string, cwd string, stdout, stderr *string) error {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Dir = cwd

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Run command \"%v\" error: %s", args, err)
	}

	if stdout != nil {
		*stdout = string(stdoutBuf.Bytes())
	}
	if stderr != nil {
		*stderr = string(stderrBuf.Bytes())
	}
	return nil
}
