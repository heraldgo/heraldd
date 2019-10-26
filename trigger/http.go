package trigger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/heraldgo/heraldd/util"
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

	tgr.ProcessFunc = func(w http.ResponseWriter, r *http.Request, body []byte) {
		bodyMap, err := util.JSONToMap(body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Request body error: %s", err)))
			return
		}

		requestChan <- bodyMap
		w.Write([]byte("Request param received and trigger activated\n"))
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
