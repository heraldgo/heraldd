package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/heraldgo/heraldd/util"
)

type exeServer struct {
	util.HTTPServer
	exeGit util.ExeGit
	secret string
}

func (s *exeServer) getOutputPath(pathOrigin string) (string, error) {
	if !filepath.IsAbs(pathOrigin) {
		return pathOrigin, nil
	}

	runDir := s.exeGit.WorkRunDir()
	runDirAbs, err := filepath.Abs(runDir)
	if err != nil {
		s.Errorf("[HeraldExeServer] Get abs path for \"%s\" error: %s", runDir, err)
		return "", err
	}

	pathRel, err := filepath.Rel(runDirAbs, pathOrigin)
	if err != nil {
		s.Errorf("[HeraldExeServer] Get rel path from \"%s\" to \"%s\" error: %s", runDirAbs, pathOrigin, err)
		return "", err
	}
	return pathRel, nil
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
	resultJSON, err := json.Marshal(result)
	if err != nil {
		s.Errorf("[HeraldExeServer] Generate json result error: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func (s *exeServer) respondMultiple(w http.ResponseWriter, result map[string]interface{}) {
	resultFiles, _ := util.GetMapParam(result, "file")

	buf := new(bytes.Buffer)
	mpWriter := multipart.NewWriter(buf)

	resultWriter, err := mpWriter.CreateFormField("result")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Create mime multipart form field error: %s", err)))
		return
	}
	resultWriter.Write(result)

	for _, fn := range files {
		fnOutput, err := s.getOutputPath(fn)
		if err != nil {
			s.Errorf("[HeraldExeServer] Get output path error: %s", err)
			continue
		}

		fd, err := os.Open(filepath.Join(s.exeGit.WorkRunDir(), fnOutput))
		if err != nil {
			s.Warnf("[HeraldExeServer] Output file open error: %s", err)
			continue
		}
		defer fd.Close()

		fWriter, err := mpWriter.CreateFormFile("file", fnOutput)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Create mime multipart form file error: %s", err)))
			return
		}

		_, err = io.Copy(fWriter, fd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Create mime multipart copy file error: %s", err)))
			return
		}
	}

	mpWriter.Close()

	w.Header().Set("Content-Type", mpWriter.FormDataContentType())
	_, err = io.Copy(w, buf)
	if err != nil {
		s.Errorf("[HeraldExeServer] Write response body error: %s", err)
	}
}

func (s *exeServer) processExecution(w http.ResponseWriter, r *http.Request, body []byte) {
	s.Infof("[HeraldExeServer] Start to execute with param: %s", string(body))

	bodyMap, err := util.JSONToMap(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Request body error: %s", err)))
		return
	}

	result := s.exeGit.Execute(bodyMap)

	if len(resultFiles) == 0 {
		s.respondSingle(w, result)
	} else {
		s.respondMultiple(w, result)
	}
}

func (s *exeServer) Run(ctx context.Context) {
	s.ValidateFunc = s.validateSignature
	s.ProcessFunc = s.processExecution

	s.HTTPServer.Run(ctx)
}

// SetLogger will set logger
func (s *exeServer) SetLogger(logger interface{}) {
	s.HTTPServer.SetLogger(logger)
	s.exeGit.SetLogger(logger)
}
