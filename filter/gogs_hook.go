package filter

import (
	"net/url"
	"strings"

	"github.com/xianghuzhao/heraldd/util"
)

// GogsHook is a filter only pass with specified repo and branch
type GogsHook struct {
	util.BaseLogger
}

// Filter will only pass with specified repo and branch
func (flt *GogsHook) Filter(triggerParam, filterParam map[string]interface{}) (map[string]interface{}, bool) {
	repository, ok := triggerParam["repository"]
	if !ok {
		flt.Logger.Errorf("[Filter(GogsHook)] Invalid gogs hook request: key \"repository\" not found")
		return nil, false
	}

	repositoryMap, ok := repository.(map[string]interface{})
	if !ok {
		flt.Logger.Errorf("[Filter(GogsHook)] Invalid gogs hook request: \"payload[repository]\" is not a map")
		return nil, false
	}

	repositoryName, _ := repositoryMap["name"]
	repoName, _ := repositoryName.(string)

	repositoryFullName, _ := repositoryMap["full_name"]
	repoFullName, _ := repositoryFullName.(string)

	repositoryCloneURL, _ := repositoryMap["clone_url"]
	cloneURL, _ := repositoryCloneURL.(string)

	ref, _ := triggerParam["ref"]
	repoRef, _ := ref.(string)
	refFrag := strings.Split(repoRef, "/")
	branch := refFrag[len(refFrag)-1]

	repoParsed, err := url.Parse(cloneURL)
	if err != nil {
		flt.Logger.Errorf("[Filter(GogsHook)] Invalid clone url: \"%s\"", cloneURL)
		return nil, false
	}
	host := strings.SplitN(repoParsed.Host, ":", 2)[0]

	gogsHost, _ := util.GetStringParam(filterParam, "gogs_host")
	gogsName, _ := util.GetStringParam(filterParam, "gogs_name")
	gogsBranch, _ := util.GetStringParam(filterParam, "gogs_branch")

	if gogsHost != "" && gogsHost != host {
		flt.Logger.Debugf("[Filter(GogsHook)] Host does not match: \"%s\"", host)
		return nil, false
	}

	if gogsBranch != "" && gogsBranch != branch {
		flt.Logger.Debugf("[Filter(GogsHook)] Branch does not match: \"%s\"", branch)
		return nil, false
	}

	if gogsName != "" {
		if strings.ContainsAny(gogsName, "/") {
			if gogsName != repoFullName {
				flt.Logger.Debugf("[Filter(GogsHook)] Name does not match: \"%s\"", repoFullName)
				return nil, false
			}
		} else {
			if gogsName != repoName {
				flt.Logger.Debugf("[Filter(GogsHook)] Full name does not match: \"%s\"", repoName)
				return nil, false
			}
		}
	}

	result := map[string]interface{}{
		"gogs_clone_url": cloneURL,
		"gogs_branch":    branch,
	}

	return result, true
}
