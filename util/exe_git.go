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

const defaultParamEnvName = "HERALD_EXECUTE_PARAM"

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

	repoPathFrags := make([]string, 0, 16)

	if host != "" {
		repoPathFrags = append(repoPathFrags, host)
	}

	urlPath = strings.TrimLeft(urlPath, "/")
	urlPath = strings.TrimSuffix(urlPath, ".git")
	repoPathFrags = append(repoPathFrags, strings.Split(urlPath, "/")...)

	return filepath.Join(repoPathFrags...)
}

func (exe *ExeGit) repoCommandPath(repo, branch, cmd string) string {
	repoPath := filepath.Join(exe.WorkDir, "gitrepo", repoRelPath(repo))

	// Update the git repository
	if stat, err := os.Stat(repoPath); os.IsNotExist(err) {
		exitCode, err := RunCmd([]string{"git", "clone", repo, repoPath}, "", nil, false, nil, nil)
		if exitCode != 0 || err != nil {
			exe.Errorf(`"git clone %s %s" error: exit(%d) err(%s)`, repo, repoPath, exitCode, err)
			return ""
		}
	} else {
		if !stat.IsDir() {
			exe.Errorf("Path for repo is not a directory: %s", repoPath)
			return ""
		}
		exitCode, err := RunCmd([]string{"git", "fetch", "--all"}, repoPath, nil, false, nil, nil)
		if exitCode != 0 || err != nil {
			exe.Errorf(`"git fetch --all" error: exit(%d) err(%s)`, exitCode, err)
			return ""
		}
	}

	if branch == "" {
		branch = "master"
	}
	exitCode, err := RunCmd([]string{"git", "reset", "--hard", "origin/" + branch}, repoPath, nil, false, nil, nil)
	if exitCode != 0 || err != nil {
		exe.Errorf(`"git reset --hard origin/%s" error: exit(%d) err(%s)`, branch, exitCode, err)
		return ""
	}
	exitCode, err = RunCmd([]string{"git", "clean", "-dfx"}, repoPath, nil, false, nil, nil)
	if exitCode != 0 || err != nil {
		exe.Warnf(`"git clean -dfx" error: exit(%d) err(%s)`, exitCode, err)
	}

	return filepath.Join(repoPath, cmd)
}

// Execute will executes script from git repo
func (exe *ExeGit) Execute(param map[string]interface{}) map[string]interface{} {
	if exe.WorkDir == "" {
		exe.Errorf("WorkDir must be specified")
		return nil
	}

	jobParam, _ := GetMapParam(param, "job_param")

	repo, _ := GetStringParam(jobParam, "repo")
	branch, _ := GetStringParam(jobParam, "branch")
	cmd, _ := GetStringParam(jobParam, "cmd")
	arg, _ := GetStringSliceParam(jobParam, "arg")
	env, _ := GetMapParam(jobParam, "env")
	paramEnvName, _ := GetStringParam(jobParam, "param_env_name")
	background, _ := GetBoolParam(jobParam, "background")
	ignoreParamEnv, _ := GetBoolParam(jobParam, "ignore_param_env")

	var finalCommand string
	if repo == "" {
		finalCommand = cmd
	} else {
		finalCommand = exe.repoCommandPath(repo, branch, cmd)
	}

	if finalCommand == "" {
		exe.Errorf("Could not execute empty command")
		return nil
	}

	var envList []string

	for k, v := range env {
		value, ok := v.(string)
		if !ok {
			continue
		}
		envList = append(envList, k+"="+value)
	}

	if !ignoreParamEnv {
		paramEnvBytes, err := json.Marshal(param)
		if err != nil {
			exe.Errorf("Generate param env failed: %s", err)
		} else {
			if paramEnvName == "" {
				paramEnvName = defaultParamEnvName
			}
			envList = append(envList, paramEnvName+"="+string(paramEnvBytes))
		}
	}

	runDir := exe.WorkRunDir()
	err := os.MkdirAll(runDir, 0755)
	if err != nil {
		exe.Errorf(`Create run directory "%s" failed: %s`, runDir, err)
	}

	fullCommand := []string{finalCommand}
	fullCommand = append(fullCommand, arg...)

	var stdout string
	exe.Debugf("Execute command: %v", fullCommand)
	exitCode, err := RunCmd(fullCommand, runDir, envList, background, &stdout, nil)
	if err != nil {
		exe.Errorf("Execute command error: %s", err)
		return nil
	}

	outputMap, err := JSONToMap([]byte(stdout))
	if err != nil {
		return map[string]interface{}{
			"output":    stdout,
			"exit_code": exitCode,
		}
	}

	outputMap["exit_code"] = exitCode

	return outputMap
}

// WorkRunDir return the run directory
func (exe *ExeGit) WorkRunDir() string {
	return filepath.Join(exe.WorkDir, "run")
}
