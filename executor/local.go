package executor

import (
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/xianghuzhao/heraldd/util"
)

// Local is an executor which runs jobs locally
type Local struct {
	util.BaseLogger
	WorkDir string
}

// Execute will run job locally
func (exe *Local) Execute(param map[string]interface{}) map[string]interface{} {
	err := os.MkdirAll(exe.WorkDir, 0755)
	if err != nil {
		exe.Logger.Errorf("[Executor(Local)] Create work directory \"%s\" failed: %s", exe.WorkDir, err)
	}

	var scriptRepo, scriptBranch, scriptCommand string
	util.GetStringParam(&scriptRepo, param, "script_repo")
	util.GetStringParam(&scriptBranch, param, "script_branch")
	util.GetStringParam(&scriptCommand, param, "script_command")

	repoParsed, err := url.Parse(scriptRepo)
	if err != nil {
		exe.Logger.Errorf("[Executor(Local)] Invalid repo name: %s", scriptRepo)
		return nil
	}

	host := strings.SplitN(repoParsed.Host, ":", 2)[0]
	repoPathFrags := []string{exe.WorkDir, host}
	urlPath := strings.TrimLeft(repoParsed.Path, "/")
	if strings.HasSuffix(urlPath, ".git") {
		urlPath = urlPath[:len(urlPath)-4]
	}
	repoPathFrags = append(repoPathFrags, strings.Split(urlPath, "/")...)
	repoPath := path.Join(repoPathFrags...)

	if stat, err := os.Stat(repoPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", scriptRepo, repoPath)
		err := cmd.Run()
		if err != nil {
			exe.Logger.Errorf("[Executor(Local)] Run \"git clone\" error: %s", err)
			return nil
		}
	} else {
		if !stat.IsDir() {
			exe.Logger.Errorf("[Executor(Local)] Path for repo is not a directory: %s", repoPath)
			return nil
		}
		cmd := exec.Command("git", "fetch", "--all")
		cmd.Dir = repoPath
		err := cmd.Run()
		if err != nil {
			exe.Logger.Errorf("[Executor(Local)] Run \"git fetch -all\" error: %s", err)
			return nil
		}
	}

	return map[string]interface{}{
		"print": true,
	}
}

// SetParam will set param from a map
func (exe *Local) SetParam(param map[string]interface{}) {
	util.GetStringParam(&exe.WorkDir, param, "work_dir")
}
