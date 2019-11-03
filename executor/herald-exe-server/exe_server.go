package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/heraldgo/heraldd/util"
)

type exeServer struct {
	util.HTTPServer
	exeGit util.ExeGit
	secret string
}

func (s *exeServer) getOutputPath(pathOrigin string) string {
	if filepath.IsAbs(pathOrigin) {
		return pathOrigin
	}
	return filepath.Join(s.exeGit.WorkRunDir(), pathOrigin)
}

func (s *exeServer) validateSignature(r *http.Request, body []byte) error {
	if r.Method != "POST" {
		return fmt.Errorf("Only POST request allowed")
	}

	sigHeader := r.Header.Get("X-Herald-Signature")
	signature, err := hex.DecodeString(sigHeader)
	if err != nil {
		return fmt.Errorf("Invalid X-Herald-Signature: %s", sigHeader)
	}
	key := []byte(s.secret)

	if !util.ValidateMAC(body, signature, key) {
		return fmt.Errorf("Signature validation Error")
	}
	return nil
}

func (s *exeServer) respondSingle(w http.ResponseWriter, result map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	delete(result, "file")
	resultJSON, err := json.Marshal(result)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error":"Generate json result error: %s"}`, err)))
	} else {
		w.Write(resultJSON)
	}
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (s *exeServer) writeResultPart(mpw *multipart.Writer, result map[string]interface{}) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, "result"))
	h.Set("Content-Type", "application/json")
	rpw, err := mpw.CreatePart(h)
	if err != nil {
		s.Errorf("Create multipart form field error: %s", err)
		return
	}

	delete(result, "file")
	resultJSON, err := json.Marshal(result)
	if err != nil {
		rpw.Write([]byte(fmt.Sprintf(`{"error":"Generate json result error: %s"}`, err)))
	} else {
		rpw.Write(resultJSON)
	}
}

func (s *exeServer) writeFilePart(mpw *multipart.Writer, name, filePath string) {
	sha256Sum, err := util.SHA256SumFile(filePath)
	if err != nil {
		s.Errorf("Get sha256 checksum error: %s", err)
		return
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"; sha256sum="%s"`,
			escapeQuotes(name), escapeQuotes(filepath.Base(filePath)),
			hex.EncodeToString(sha256Sum)))
	h.Set("Content-Type", "application/octet-stream")

	fpw, err := mpw.CreatePart(h)
	if err != nil {
		s.Errorf("Create multipart form field error: %s", err)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		s.Errorf("Output file open error: %s", err)
		return
	}
	defer f.Close()

	_, err = io.Copy(fpw, f)
	if err != nil {
		s.Errorf("Multipart copy file error: %s", err)
		return
	}
}

func (s *exeServer) respondMultiple(w http.ResponseWriter, result map[string]interface{}) {
	resultFiles, _ := util.GetMapParam(result, "file")

	mpWriter := multipart.NewWriter(w)

	w.Header().Set("Content-Type", mpWriter.FormDataContentType())

	s.writeResultPart(mpWriter, result)

	for name, filePath := range resultFiles {
		fp, ok := filePath.(string)
		if !ok {
			s.Warnf("File value must be string of file path: %v", filePath)
			continue
		}
		fnOutput := s.getOutputPath(fp)
		s.writeFilePart(mpWriter, name, fnOutput)
	}

	mpWriter.Close()
}

func (s *exeServer) processExecution(w http.ResponseWriter, r *http.Request, body []byte) {
	s.Infof("Start to execute with param: %s", string(body))

	bodyMap, err := util.JSONToMap(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Request body error: %s", err)))
		return
	}

	result := s.exeGit.Execute(bodyMap)
	s.Debugf("Execute result: %s", result)

	fileMap, _ := result["file"].(map[string]interface{})

	if len(fileMap) == 0 {
		s.respondSingle(w, result)
	} else {
		s.respondMultiple(w, result)
	}
}

func (s *exeServer) Run(ctx context.Context) {
	s.ValidateFunc = s.validateSignature
	s.ProcessFunc = s.processExecution

	s.Start()
	defer s.Stop()

	<-ctx.Done()
}

// SetLogger will set logger for both HTTPServer and exeGit
func (s *exeServer) SetLogger(logger interface{}) {
	s.HTTPServer.SetLogger(logger)
	s.exeGit.SetLogger(logger)
}

// SetLoggerPrefix will set logger prefix for both HTTPServer and exeGit
func (s *exeServer) SetLoggerPrefix(prefix string) {
	s.HTTPServer.SetLoggerPrefix(prefix)
	s.exeGit.SetLoggerPrefix(prefix)
}
