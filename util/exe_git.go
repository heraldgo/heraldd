package util

import (
	"encoding/json"
	"net/url"
	"os"
	"path"
	"strings"
)

// ExeGit executes script from git repository
type ExeGit struct {
	BaseLogger
	WorkDir string
}

// Execute will executes script from git repo
func (exe *ExeGit) Execute(param map[string]interface{}) map[string]interface{} {
	err := os.MkdirAll(exe.WorkDir, 0755)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] Create work directory \"%s\" failed: %s", exe.WorkDir, err)
	}

	scriptRepo, _ := GetStringParam(param, "script_repo")
	scriptBranch, _ := GetStringParam(param, "script_branch")
	scriptCommand, _ := GetStringParam(param, "command")

	repoParsed, err := url.Parse(scriptRepo)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] Invalid repo name: %s", scriptRepo)
		return nil
	}

	host := strings.SplitN(repoParsed.Host, ":", 2)[0]
	repoPathFrags := []string{exe.WorkDir, "repo", host}
	urlPath := strings.TrimLeft(repoParsed.Path, "/")
	if strings.HasSuffix(urlPath, ".git") {
		urlPath = urlPath[:len(urlPath)-4]
	}
	repoPathFrags = append(repoPathFrags, strings.Split(urlPath, "/")...)
	repoPath := path.Join(repoPathFrags...)

	if stat, err := os.Stat(repoPath); os.IsNotExist(err) {
		err := RunCmd([]string{"git", "clone", scriptRepo, repoPath}, "", nil, nil)
		if err != nil {
			exe.Errorf("[Util(ExeGit)] %s", err)
			return nil
		}
	} else {
		if !stat.IsDir() {
			exe.Errorf("[Util(ExeGit)] Path for repo is not a directory: %s", repoPath)
			return nil
		}
		err := RunCmd([]string{"git", "fetch", "--all"}, repoPath, nil, nil)
		if err != nil {
			exe.Errorf("[Util(ExeGit)] %s", err)
			return nil
		}
	}

	if scriptBranch == "" {
		scriptBranch = "master"
	}
	err = RunCmd([]string{"git", "checkout", "refs/remotes/origin/" + scriptBranch}, repoPath, nil, nil)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] %s", err)
		return nil
	}

	runDir := path.Join(exe.WorkDir, "run")
	err = os.MkdirAll(runDir, 0755)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] Create run directory \"%s\" failed: %s", runDir, err)
	}

	var paramArg string
	paramArgBytes, err := json.Marshal(param)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] Generate param argument failed: %s", err)
	} else {
		paramArg = string(paramArgBytes)
	}

	var stdout string
	err = RunCmd([]string{path.Join(repoPath, scriptCommand), paramArg}, runDir, &stdout, nil)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] %s", err)
		return nil
	}

	var outputResult interface{}
	err = json.Unmarshal([]byte(stdout), &outputResult)
	outputMap, ok := outputResult.(map[string]interface{})
	if err != nil || !ok {
		return map[string]interface{}{
			"output": stdout,
		}
	}

	return outputMap
}
