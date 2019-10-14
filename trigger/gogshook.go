package trigger

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/xianghuzhao/heraldd/util"
)

// GogsHook is a trigger which will listen to http request
type GogsHook struct {
	util.HTTPServer
	Secret string
}

func (tgr *GogsHook) validateGogs(r *http.Request, body []byte) error {
	if r.Method != "POST" {
		return fmt.Errorf("Only POST request allowed")
	}

	sigHeader := r.Header.Get("X-Gogs-Signature")
	signature, err := hex.DecodeString(sigHeader)
	if err != nil {
		return fmt.Errorf("Invalid X-Gogs-Signature: %s", sigHeader)
	}
	key := []byte(tgr.Secret)

	if !util.ValidateMAC(body, signature, key) {
		return fmt.Errorf("Signature validation Error")
	}
	return nil
}

// Run the GogsHook trigger
func (tgr *GogsHook) Run(ctx context.Context, param chan map[string]interface{}) {
	tgr.ValidateFunc = tgr.validateGogs

	requestChan := make(chan map[string]interface{})

	tgr.ProcessFunc = func(w http.ResponseWriter, reqParam map[string]interface{}) error {
		requestChan <- reqParam
		w.Write([]byte("Gogs param received and trigger activated\n"))
		return nil
	}

	go func() {
		tgr.HTTPServer.Run(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case reqParam := <-requestChan:
			param <- reqParam
		}
	}
}

// SetParam will set param from a map
func (tgr *GogsHook) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&tgr.UnixSocket, param, "unix_socket")
	util.UpdateStringParam(&tgr.Host, param, "host")
	util.UpdateIntParam(&tgr.Port, param, "port")
	util.UpdateStringParam(&tgr.Secret, param, "secret")
}
