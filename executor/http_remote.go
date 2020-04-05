package executor

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	DataDir string
}

func (exe *HTTPRemote) processJSONPart(result map[string]interface{}, reader io.Reader) {
	body, _ := ioutil.ReadAll(reader)
	bodyMap, err := util.JSONToMap(body)
	if err != nil {
		result["response"] = string(body)
	} else {
		delete(bodyMap, "file")
		util.MergeMapParam(result, bodyMap)
	}
}

func (exe *HTTPRemote) processFilePart(result map[string]interface{}, part *multipart.Part, exeID string) {
	contentDisposition := part.Header.Get("Content-Disposition")
	_, cdParams, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		exe.Errorf("Parse Content-Disposition error: %s", err)
		return
	}

	name := cdParams["name"]
	if name == "" {
		exe.Errorf("Multipart name not found")
		return
	}

	filename := cdParams["filename"]
	if filename == "" {
		exe.Errorf("Multipart filename not found")
		return
	}

	fileDir := filepath.Join(exe.DataDir, exeID, name)
	filePath := filepath.Join(fileDir, filename)

	err = os.MkdirAll(fileDir, 0755)
	if err != nil {
		exe.Errorf(`Create data file directory "%s" failed: %s`, fileDir, err)
		return
	}

	func() {
		out, err := os.Create(filePath)
		if err != nil {
			exe.Errorf(`Create file "%s" error: %s`, filePath, err)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, part)
		if err != nil {
			exe.Errorf(`Write file "%s" error: %s`, filePath, err)
			return
		}
	}()

	sha256SumString := cdParams["sha256sum"]
	if sha256SumString != "" {
		sha256Sum, err := hex.DecodeString(sha256SumString)
		if err != nil {
			exe.Errorf(`Decode checksum "%s" error: %s`, sha256SumString, err)
			return
		}

		fileCheckSum, err := util.SHA256SumFile(filePath)
		if err != nil {
			exe.Errorf(`Get file "%s" checksum error: %s`, filePath, err)
			return
		}

		exe.Debugf(`File "%s" sha256 checksum "%x": expect "%s"`, filePath, fileCheckSum, sha256SumString)
		if !bytes.Equal(sha256Sum, fileCheckSum) {
			exe.Errorf(`File sha256 checksum does not match for "%s", expecting "%s"`, filePath, sha256SumString)
			return
		}
	}

	_, ok := result["file"].(map[string]interface{})
	if !ok {
		result["file"] = make(map[string]interface{})
	}
	mapFile, _ := result["file"].(map[string]interface{})
	mapFile[name] = filePath
}

func (exe *HTTPRemote) processMultiPart(result map[string]interface{}, reader io.Reader, boundary, exeID string) {
	mpReader := multipart.NewReader(reader, boundary)
	for {
		part, err := mpReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			exe.Errorf("Read multipart error: %s", err)
			return
		}

		contentType := part.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			exe.Errorf("Parse part media type error: %s", err)
			continue
		}

		switch mediaType {
		case "application/json":
			exe.processJSONPart(result, part)
		case "application/octet-stream":
			exe.processFilePart(result, part, exeID)
		}
	}
}

// Execute will run job on the remote server
func (exe *HTTPRemote) Execute(param map[string]interface{}) (map[string]interface{}, error) {
	exeID, _ := util.GetStringParam(param, "id")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		exe.Errorf("Generate json param error: %s", err)
		return nil, errors.New("Generate json param error")
	}

	signatureBytes := util.CalculateMAC(paramJSON, []byte(exe.Secret))
	signature := hex.EncodeToString(signatureBytes)

	req, err := http.NewRequest("POST", exe.Host, bytes.NewBuffer(paramJSON))
	if err != nil {
		exe.Errorf("Create request failed: %s", err)
		return nil, errors.New("Create request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Herald-Signature", signature)

	exe.Infof("Start to connect to: %s", exe.Host)

	client := &http.Client{
		Timeout: exe.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		exe.Errorf("Remote execution request failed: %s", err)
		return nil, errors.New("Remote execution request failed")
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	exe.Debugf("Response status: %s", resp.Status)
	exe.Debugf("Response content type: %s", contentType)

	if resp.StatusCode != http.StatusOK {
		exe.Errorf("Http status not OK: %s", resp.Status)
		body, _ := ioutil.ReadAll(resp.Body)
		exe.Errorf("Remote error: %s", string(body))
		return nil, fmt.Errorf(`Http status %d: "%s"`, resp.StatusCode, string(body))
	}

	mediaType, mtParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		exe.Errorf("Parse media type error: %s", err)
		return nil, errors.New("Parse media type error")
	}

	result := make(map[string]interface{})

	exe.Debugf("Parsed context type: %s", mediaType)
	result["context_type"] = mediaType

	if mediaType == "application/json" {
		exe.processJSONPart(result, resp.Body)
	} else if strings.HasPrefix(mediaType, "multipart/") {
		exe.processMultiPart(result, resp.Body, mtParams["boundary"], exeID)
	} else {
		exe.Errorf("Unknown media type: %s", mediaType)
		body, _ := ioutil.ReadAll(resp.Body)
		result["response"] = string(body)
		return result, errors.New("Unknown media type")
	}

	exitCodeFloat, err := util.GetFloatParam(result, "exit_code")
	exitCode := int(exitCodeFloat)
	if exitCode != 0 {
		return result, fmt.Errorf("Command failed with code %d", exitCode)
	}

	return result, nil
}

func newExecutorHTTPRemote(param map[string]interface{}) interface{} {
	host, _ := util.GetStringParam(param, "host")
	timeout, _ := util.GetIntParam(param, "timeout")
	secret, _ := util.GetStringParam(param, "secret")
	dataDir, _ := util.GetStringParam(param, "data_dir")

	exe := &HTTPRemote{
		Host:    host,
		Timeout: time.Duration(timeout) * time.Second,
		Secret:  secret,
		DataDir: dataDir,
	}
	return exe
}
