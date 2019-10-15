package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/xianghuzhao/heraldd/util"
)

type exeServer struct {
	util.HTTPServer
	secret string
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

func (s *exeServer) processExecution(w http.ResponseWriter, reqParam map[string]interface{}) {
	s.Infof("Start to execute with param: %v", reqParam)
	w.Write([]byte("Exe server param received:\n"))
	w.Write([]byte(fmt.Sprintf("%v\n", reqParam)))
}

func (s *exeServer) Run(ctx context.Context) {
	s.ValidateFunc = s.validateSignature
	s.ProcessFunc = s.processExecution

	s.HTTPServer.Run(ctx)
}
