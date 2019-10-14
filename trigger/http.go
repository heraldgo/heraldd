package trigger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xianghuzhao/heraldd/util"
)

// HTTP is a trigger which will listen to http request
type HTTP struct {
	util.HTTPServer
}

func (tgr *HTTP) validateMethod(r *http.Request, body []byte) error {
	if r.Method != "POST" {
		return fmt.Errorf("Only POST request allowed")
	}
	return nil
}

// Run the HTTP trigger
func (tgr *HTTP) Run(ctx context.Context, param chan map[string]interface{}) {
	tgr.ValidateFunc = func(r *http.Request, body []byte) error {
		if r.Method != "POST" {
			return fmt.Errorf("Only POST request allowed")
		}
		return nil
	}

	requestChan := make(chan map[string]interface{})

	tgr.ProcessFunc = func(w http.ResponseWriter, reqParam map[string]interface{}) error {
		requestChan <- reqParam
		w.Write([]byte("Request param received and trigger activated\n"))
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
func (tgr *HTTP) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&tgr.UnixSocket, param, "unix_socket")
	util.UpdateStringParam(&tgr.Host, param, "host")
	util.UpdateIntParam(&tgr.Port, param, "port")
}
