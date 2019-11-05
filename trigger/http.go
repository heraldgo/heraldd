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

// Run the HTTP trigger
func (tgr *HTTP) Run(ctx context.Context, sendParam func(map[string]interface{})) {
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

		select {
		case <-ctx.Done():
			return
		case requestChan <- bodyMap:
		}

		w.Write([]byte("Request param received and trigger activated\n"))
	}

	tgr.Start()
	defer tgr.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case reqParam := <-requestChan:
			sendParam(reqParam)
		}
	}
}

func newTriggerHTTP(param map[string]interface{}) interface{} {
	unixSocket, _ := util.GetStringParam(param, "unix_socket")
	host, _ := util.GetStringParam(param, "host")
	port, _ := util.GetIntParam(param, "port")

	if port == 0 && unixSocket == "" {
		host = "127.0.0.1"
		port = 8123
	}

	return &HTTP{
		HTTPServer: util.HTTPServer{
			UnixSocket: unixSocket,
			Host:       host,
			Port:       port,
		},
	}
}
