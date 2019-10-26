package executor

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heraldgo/herald"

	"github.com/heraldgo/heraldd/util"
)

// ExeClient is an executor which runs jobs on the remote server
type ExeClient struct {
	util.BaseLogger
	Host    string
	Timeout time.Duration
	Secret  string
	WorkDir string
}

func (exe *ExeClient) processJSONPart(result map[string]interface{}, reader io.Reader) {
	body, _ := ioutil.ReadAll(reader)
	bodyMap, err := util.JSONToMap(body)
	if err != nil {
		result["response"] = string(body)
	} else {
		_, ok := bodyMap["file"]
		if ok {
			delete(bodyMap, "file")
		}
		herald.UpdateMapParam(result, bodyMap)
	}
}

func (exe *ExeClient) processFilePart(result map[string]interface{}, part *multipart.Part) {
	filename := part.FileName()
	if filename == "" {
		exe.Errorf("[Executor(ExeClient)] Multipart filename not found")
		return
	}
	fn := filepath.Join(exe.WorkDir, filename)
	out, err := os.Create(fn)
	if err != nil {
		exe.Errorf("[Executor(ExeClient)] Create file \"%s\" error: %s", fn, err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, part)
	if err != nil {
		exe.Errorf("[Executor(ExeClient)] Write file error: %s", err)
		return
	}

	resultFile, ok := result["file"]
	if !ok {
		result["file"] = make([]interface{}, 0)
	}
	resultFileSlice, ok := resultFile.([]interface{})
	if !ok {
		exe.Errorf("[Executor(ExeClient)] Result file is not array")
		return
	}
	resultFileSlice = append(resultFileSlice, fn)
}

// Execute will run job on the remote server
func (exe *ExeClient) Execute(param map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	paramJSON, err := json.Marshal(param)
	if err != nil {
		exe.Errorf("[Executor(ExeClient)] Generate json param error: %s", err)
		return nil
	}

	signatureBytes := util.CalculateMAC(paramJSON, []byte(exe.Secret))
	signature := hex.EncodeToString(signatureBytes)

	req, err := http.NewRequest("POST", exe.Host, bytes.NewBuffer(paramJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Herald-Signature", signature)

	exe.Infof("[Executor(ExeClient)] Start to connect to: %s", exe.Host)

	client := &http.Client{
		Timeout: exe.Timeout * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		exe.Errorf("[Executor(ExeClient)] Remote execution request failed: %s", err)
		return nil
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	exe.Debugf("[Executor(ExeClient)] Response status: %s", resp.Status)
	exe.Debugf("[Executor(ExeClient)] Response content type: %s", contentType)

	if resp.StatusCode != http.StatusOK {
		exe.Errorf("[Executor(ExeClient)] Http status not OK: %s", resp.Status)
		return nil
	}

	mediaType, mtParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		exe.Errorf("[Executor(ExeClient)] Parse media type error: %s", err)
		return nil
	}

	exe.Debugf("[Executor(ExeClient)] Context type: %s", mediaType)
	result["context_type"] = mediaType

	if mediaType == "application/json" {
		exe.processJSONPart(result, resp.Body)
	} else if strings.HasPrefix(mediaType, "multipart/") {
		mpReader := multipart.NewReader(resp.Body, mtParams["boundary"])
		for {
			part, err := mpReader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				exe.Errorf("[Executor(ExeClient)] Read multipart error: %s", err)
				return nil
			}

			contentDisposition := part.Header.Get("Content-Disposition")
			_, cdParams, err := mime.ParseMediaType(contentDisposition)
			if err != nil {
				exe.Errorf("[Executor(ExeClient)] Parse Content-Disposition error: %s", err)
				continue
			}

			cdName, _ := cdParams["name"]
			if cdName == "result" {
				exe.processJSONPart(result, part)
			} else if cdName == "file" {
				exe.processFilePart(result, part)
			}
		}
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		result["response"] = string(body)
	}

	return result
}

// SetParam will set param from a map
func (exe *ExeClient) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.Host, param, "host")
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
	util.UpdateStringParam(&exe.Secret, param, "secret")

	timeout, err := util.GetIntParam(param, "timeout")
	if err != nil {
		exe.Timeout = time.Duration(timeout) * time.Second
	}
}
