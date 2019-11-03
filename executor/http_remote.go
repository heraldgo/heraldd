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

	"github.com/heraldgo/heraldd/util"
)

// HTTPRemote is an executor which runs jobs on the remote server
type HTTPRemote struct {
	util.BaseLogger
	Host    string
	Timeout time.Duration
	Secret  string
	WorkDir string
}

func (exe *HTTPRemote) processJSONPart(result map[string]interface{}, reader io.Reader) {
	body, _ := ioutil.ReadAll(reader)
	bodyMap, err := util.JSONToMap(body)
	if err != nil {
		result["response"] = string(body)
	} else {
		_, ok := bodyMap["file"]
		if ok {
			delete(bodyMap, "file")
		}
		util.MergeMapParam(result, bodyMap)
	}
}

func (exe *HTTPRemote) processFilePart(result map[string]interface{}, part *multipart.Part) {
	filename := part.FileName()
	if filename == "" {
		exe.Errorf("Multipart filename not found")
		return
	}
	fn := filepath.Join(exe.WorkDir, filename)
	out, err := os.Create(fn)
	if err != nil {
		exe.Errorf(`Create file "%s" error: %s`, fn, err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, part)
	if err != nil {
		exe.Errorf("Write file error: %s", err)
		return
	}

	resultFile, ok := result["file"]
	if !ok {
		result["file"] = make([]interface{}, 0)
	}
	resultFileSlice, ok := resultFile.([]interface{})
	if !ok {
		exe.Errorf("Result file is not array")
		return
	}
	resultFileSlice = append(resultFileSlice, fn)
}

// Execute will run job on the remote server
func (exe *HTTPRemote) Execute(param map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	paramJSON, err := json.Marshal(param)
	if err != nil {
		exe.Errorf("Generate json param error: %s", err)
		return nil
	}

	signatureBytes := util.CalculateMAC(paramJSON, []byte(exe.Secret))
	signature := hex.EncodeToString(signatureBytes)

	req, err := http.NewRequest("POST", exe.Host, bytes.NewBuffer(paramJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Herald-Signature", signature)

	exe.Infof("Start to connect to: %s", exe.Host)

	client := &http.Client{
		Timeout: exe.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		exe.Errorf("Remote execution request failed: %s", err)
		return nil
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	exe.Debugf("Response status: %s", resp.Status)
	exe.Debugf("Response content type: %s", contentType)

	if resp.StatusCode != http.StatusOK {
		exe.Errorf("Http status not OK: %s", resp.Status)
		return nil
	}

	mediaType, mtParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		exe.Errorf("Parse media type error: %s", err)
		return nil
	}

	exe.Debugf("Context type: %s", mediaType)
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
				exe.Errorf("Read multipart error: %s", err)
				return nil
			}

			contentDisposition := part.Header.Get("Content-Disposition")
			_, cdParams, err := mime.ParseMediaType(contentDisposition)
			if err != nil {
				exe.Errorf("Parse Content-Disposition error: %s", err)
				continue
			}

			cdName := cdParams["name"]
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

func newExecutorHTTPRemote(param map[string]interface{}) interface{} {
	host, _ := util.GetStringParam(param, "host")
	timeout, _ := util.GetIntParam(param, "timeout")
	secret, _ := util.GetStringParam(param, "secret")
	workDir, _ := util.GetStringParam(param, "work_dir")

	exe := &HTTPRemote{
		Host:    host,
		Timeout: time.Duration(timeout) * time.Second,
		Secret:  secret,
		WorkDir: workDir,
	}
	return exe
}
