package util

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// ExeGit executes script from git repository
type ExeGit struct {
	BaseLogger
	WorkDir string
}

const (
	defaultParamEnvName = "HERALD_EXECUTE_PARAM"
	anonymousRemoteName = "herald-executor-remote-anonymous"
)

var sshDefaultKeyFiles = [...]string{"id_dsa", "id_ecdsa", "id_ed25519", "id_rsa"}

func getUserPass(endpoint *transport.Endpoint, username, password *string) {
	if *username == "" {
		*username = endpoint.User
	}

	if *password == "" {
		*password = endpoint.Password
	}
	endpoint.Password = ""
}

func (exe *ExeGit) getHTTPAuth(endpoint *transport.Endpoint, username, password string) transport.AuthMethod {
	getUserPass(endpoint, &username, &password)

	if username != "" || password != "" {
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	return nil
}

func findDefaultSSHKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	for _, keyFile := range sshDefaultKeyFiles {
		keyPath := filepath.Join(home, ".ssh", keyFile)
		if FileExists(keyPath) {
			return keyPath
		}
	}
	return ""
}

func (exe *ExeGit) getSSHAuth(endpoint *transport.Endpoint, username, password string, key []byte, keyFile, keyPassword string) transport.AuthMethod {
	getUserPass(endpoint, &username, &password)

	if password != "" {
		return &ssh.Password{
			User:     username,
			Password: password,
		}
	}

	if len(key) != 0 {
		auth, err := ssh.NewPublicKeys(username, key, keyPassword)
		if err != nil {
			exe.Errorf("Get ssh key error: %s", err)
			return nil
		}
		return auth
	}

	if keyFile == "" {
		keyFile = findDefaultSSHKey()
	}

	if keyFile != "" {
		exe.Infof("Use ssh key file: %s", keyFile)
		auth, err := ssh.NewPublicKeysFromFile(username, keyFile, keyPassword)
		if err != nil {
			exe.Errorf(`Get ssh key from file "%s" error: %s`, keyFile, err)
			return nil
		}
		return auth
	}

	return nil
}

func (exe *ExeGit) getParam(repo, username, password, key, keyFile, keyPassword string) (string, string, transport.AuthMethod, error) {
	var auth transport.AuthMethod

	endpoint, err := transport.NewEndpoint(repo)
	if err != nil {
		return "", "", nil, err
	}

	if endpoint.Protocol == "http" || endpoint.Protocol == "https" {
		auth = exe.getHTTPAuth(endpoint, username, password)
	} else if endpoint.Protocol == "ssh" {
		auth = exe.getSSHAuth(endpoint, username, password, []byte(key), keyFile, keyPassword)
	}

	repoPathFrags := make([]string, 0, 16)
	repoPathFrags = append(repoPathFrags, endpoint.Host)
	urlPath := strings.TrimLeft(endpoint.Path, "/")
	urlPath = strings.TrimSuffix(urlPath, ".git")
	repoPathFrags = append(repoPathFrags, strings.Split(urlPath, "/")...)
	repoPath := filepath.Join(repoPathFrags...)

	return repoPath, endpoint.String(), auth, nil
}

func (exe *ExeGit) getRepo(repoDir, url string, auth transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainOpen(repoDir)

	if err != nil {
		if err != git.ErrRepositoryNotExists {
			return nil, err
		}

		repo, err = git.PlainInit(repoDir, false)
		if err != nil {
			return nil, err
		}
	}

	remote, err := repo.CreateRemoteAnonymous(&config.RemoteConfig{
		Name:  "anonymous",
		URLs:  []string{url},
		Fetch: []config.RefSpec{"+refs/heads/*:refs/remotes/" + anonymousRemoteName + "/*"},
	})
	if err != nil {
		return nil, err
	}

	err = remote.Fetch(&git.FetchOptions{
		Auth:  auth,
		Force: true,
	})

	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return nil, err
		}
		exe.Infof("Repo already up to date")
	}

	return repo, nil
}

func (exe *ExeGit) loadBranch(repo *git.Repository, branch string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	commit, err := repo.ResolveRevision(plumbing.Revision("refs/remotes/" + anonymousRemoteName + "/" + branch))
	if err != nil {
		exe.Debugf("Find remote branch failed, try to find a tag")
		commit, err = repo.ResolveRevision(plumbing.Revision("refs/tags/" + branch))
		if err != nil {
			exe.Errorf("Find branch/tag failed: %s", err)
			return errors.New("Find branch/tag failed")
		}
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash:  *commit,
		Force: true,
	})
	if err != nil {
		exe.Errorf("Repo checkout error:", err)
		return errors.New("Repo checkout error")
	}

	err = worktree.Reset(&git.ResetOptions{
		Commit: *commit,
		Mode:   git.HardReset,
	})
	if err != nil {
		exe.Errorf("Repo reset error:", err)
		return errors.New("Repo reset error")
	}

	err = worktree.Clean(&git.CleanOptions{
		Dir: true,
	})
	if err != nil {
		exe.Errorf("Repo clean error:", err)
		return errors.New("Repo clean error")
	}

	return nil
}

func (exe *ExeGit) loadRepo(repo, username, password, sshKey, sshKeyFile, sshKeyPassword, branch string) (string, error) {
	repoDir, url, auth, err := exe.getParam(repo, username, password, sshKey, sshKeyFile, sshKeyPassword)
	if err != nil {
		exe.Errorf("Get repo param error: %s", err)
		return "", errors.New("Get repo param error")
	}

	repoDir = filepath.Join(exe.WorkRepoDir(), repoDir)

	exe.Debugf("Repo dir: %s", repoDir)
	exe.Debugf("Repo URL: %s", url)
	if auth != nil {
		exe.Debugf("Repo auth: %s", auth.Name())
	}

	gitRepo, err := exe.getRepo(repoDir, url, auth)
	if err != nil {
		exe.Errorf("Get repo error:", err)
		return "", errors.New("Get repo error")
	}

	err = exe.loadBranch(gitRepo, branch)
	if err != nil {
		exe.Errorf("Load branch or tag \"%s\" error: %s", branch, err)
		return "", errors.New("Load branch error")
	}

	return repoDir, nil
}

// Execute will executes script from git repo
func (exe *ExeGit) Execute(param map[string]interface{}) (map[string]interface{}, error) {
	if exe.WorkDir == "" {
		exe.Errorf("WorkDir must be specified")
		return nil, errors.New("WorkDir must be specified")
	}

	jobParam, _ := GetMapParam(param, "job_param")

	repo, _ := GetStringParam(jobParam, "repo")
	username, _ := GetStringParam(jobParam, "username")
	password, _ := GetStringParam(jobParam, "password")
	sshKey, _ := GetStringParam(jobParam, "ssh_key")
	sshKeyFile, _ := GetStringParam(jobParam, "ssh_key_file")
	sshKeyPassword, _ := GetStringParam(jobParam, "ssh_key_password")
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
		if branch == "" {
			branch = "master"
		}
		repoPath, err := exe.loadRepo(repo, username, password, sshKey, sshKeyFile, sshKeyPassword, branch)
		if err != nil {
			exe.Errorf("Load git repository failed: %s", err)
			return nil, errors.New("Load git repository failed")
		}
		finalCommand = filepath.Join(repoPath, cmd)
	}

	if finalCommand == "" {
		exe.Errorf("Could not execute empty command")
		return nil, errors.New("Could not execute empty command")
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
			return nil, errors.New("Generate param env failed")
		}
		if paramEnvName == "" {
			paramEnvName = defaultParamEnvName
		}
		envList = append(envList, paramEnvName+"="+string(paramEnvBytes))
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
		return nil, errors.New("Execute command error")
	}

	result, err := JSONToMap([]byte(stdout))
	if err != nil {
		return map[string]interface{}{
			"output":    stdout,
			"exit_code": exitCode,
		}, nil
	}

	result["exit_code"] = exitCode

	return result, nil
}

// WorkRepoDir return the run directory
func (exe *ExeGit) WorkRepoDir() string {
	return filepath.Join(exe.WorkDir, "gitrepo")
}

// WorkRunDir return the run directory
func (exe *ExeGit) WorkRunDir() string {
	return filepath.Join(exe.WorkDir, "run")
}
