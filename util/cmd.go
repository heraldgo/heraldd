package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// RunCmd will open the sub process
func RunCmd(args []string, cwd string, env []string, background bool, stdout, stderr *string) (int, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	cmd.Dir = cwd

	err := cmd.Start()
	if err != nil {
		return -1, fmt.Errorf(`Start command "%v" error: %s`, args, err)
	}

	if background {
		return 0, nil
	}

	var exitCode int

	err = cmd.Wait()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			exitCode = exitError.ExitCode()
		} else {
			return -1, fmt.Errorf(`Run command "%v" error: %s`, args, err)
		}
	}

	if stdout != nil {
		*stdout = stdoutBuf.String()
	}
	if stderr != nil {
		*stderr = stderrBuf.String()
	}

	return exitCode, nil
}
