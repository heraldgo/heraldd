package util

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ExeGit executes script from git repository
type ExeGit struct {
	BaseLogger
	WorkDir string
}

func repoRelPath(u string) string {
	var host, urlPath string

	schemaSep := strings.Index(u, "://")
	if schemaSep == -1 {
		// git@github.com:aaa/bbb or example.com:jjj/kkk
		hostStart := strings.Index(u, "@")
		pathStart := strings.Index(u, ":")
		if pathStart == -1 {
			urlPath = u[hostStart+1:]
		} else {
			host = u[hostStart+1 : pathStart]
			urlPath = u[pathStart+1:]
		}
	} else {
		repoParsed, err := url.Parse(u)
		if err != nil {
			urlPath = u[schemaSep+3:]
		} else {
			host = strings.SplitN(repoParsed.Host, ":", 2)[0]
			urlPath = repoParsed.Path
		}
	}

	var repoPathFrags []string

	if host != "" {
		repoPathFrags = append(repoPathFrags, host)
	}

	urlPath = strings.TrimLeft(urlPath, "/")
	urlPath = strings.TrimSuffix(urlPath, ".git")
	repoPathFrags = append(repoPathFrags, strings.Split(urlPath, "/")...)

	return filepath.Join(repoPathFrags...)
}

// Execute will executes script from git repo
func (exe *ExeGit) Execute(param map[string]interface{}) map[string]interface{} {
	err := os.MkdirAll(exe.WorkDir, 0755)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] Create work directory \"%s\" failed: %s", exe.WorkDir, err)
	}

	jobParam, _ := GetMapParam(param, "job_param")

	scriptRepo, _ := GetStringParam(jobParam, "script_repo")
	scriptBranch, _ := GetStringParam(jobParam, "script_branch")
	scriptCommand, _ := GetStringParam(jobParam, "command")

	var finalCommand string
	if scriptRepo == "" {
		finalCommand = scriptCommand
	} else {
		repoPath := filepath.Join(exe.WorkDir, "repo", repoRelPath(scriptRepo))

		// Update the git repository
		if stat, err := os.Stat(repoPath); os.IsNotExist(err) {
			err := RunCmd([]string{"git", "clone", scriptRepo, repoPath}, "", nil, nil)
			if err != nil {
				exe.Errorf("[Util(ExeGit)] \"git clone\" error: %s", err)
				return nil
			}
		} else {
			if !stat.IsDir() {
				exe.Errorf("[Util(ExeGit)] Path for repo is not a directory: %s", repoPath)
				return nil
			}
			err := RunCmd([]string{"git", "fetch", "--all"}, repoPath, nil, nil)
			if err != nil {
				exe.Errorf("[Util(ExeGit)] \"git fetch -all\" error: %s", err)
				return nil
			}
		}

		if scriptBranch == "" {
			scriptBranch = "master"
		}
		err = RunCmd([]string{"git", "reset", "--hard", "origin/" + scriptBranch}, repoPath, nil, nil)
		if err != nil {
			exe.Errorf("[Util(ExeGit)] \"git reset --hard\" error: %s", err)
			return nil
		}
		err = RunCmd([]string{"git", "clean", "-dfx"}, repoPath, nil, nil)
		if err != nil {
			exe.Warnf("[Util(ExeGit)] \"git clean -dfx\" error: %s", err)
		}

		finalCommand = filepath.Join(repoPath, scriptCommand)
	}

	runDir := exe.WorkRunDir()
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
	err = RunCmd([]string{finalCommand, paramArg}, runDir, &stdout, nil)
	if err != nil {
		exe.Errorf("[Util(ExeGit)] Execute script command error: %s", err)
		return nil
	}

	outputMap, err := JSONToMap([]byte(stdout))
	if err != nil {
		return map[string]interface{}{
			"output": stdout,
		}
	}

	return map[string]interface{}{
		"result": outputMap,
	}
}

// WorkRunDir return the run directory
func (exe *ExeGit) WorkRunDir() string {
	return filepath.Join(exe.WorkDir, "run")
}
