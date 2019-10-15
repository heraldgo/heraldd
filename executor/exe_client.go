package executor

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/xianghuzhao/heraldd/util"
)

// ExeClient is an executor which runs jobs on the remote server
type ExeClient struct {
	util.BaseLogger
	Host     string
	Secret   string
	WorkDir  string
	ExtraMap map[string]interface{}
}

// Execute will run job on the remote server
func (exe *ExeClient) Execute(param map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range exe.ExtraMap {
		result[k] = v
	}

	paramJSON, err := json.Marshal(param)

	signatureBytes := util.CalculateMAC(paramJSON, []byte(exe.Secret))
	signature := hex.EncodeToString(signatureBytes)

	req, err := http.NewRequest("POST", exe.Host, bytes.NewBuffer(paramJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Herald-Signature", signature)

	exe.Infof("[Executor(ExeClient)] Start to connect to: %s", exe.Host)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		exe.Errorf("[Executor(ExeClient)] Remote execution request failed: %s", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		exe.Errorf("[Executor(ExeClient)] Http status: %s", resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var bodyJSON interface{}
	err = json.Unmarshal(body, &bodyJSON)
	bodyMap, ok := bodyJSON.(map[string]interface{})
	if err != nil || !ok {
		result["output"] = string(body)
	} else {
		for k, v := range bodyMap {
			result[k] = v
		}
	}

	return result
}

// SetParam will set param from a map
func (exe *ExeClient) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.Host, param, "host")
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
	util.UpdateStringParam(&exe.Secret, param, "secret")
	util.UpdateMapParam(&exe.ExtraMap, param, "extra_map")
}
