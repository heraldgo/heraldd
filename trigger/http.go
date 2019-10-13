package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/xianghuzhao/heraldd/util"
)

// HTTP is a trigger which will listen to http request
type HTTP struct {
	util.BaseLogger
	UnixSocket   string
	Host         string
	Port         int
	validateFunc func(*http.Request, []byte) error
	requestParam chan map[string]interface{}
}

func (tgr *HTTP) validateMethod(r *http.Request, body []byte) error {
	if r.Method != "POST" {
		return fmt.Errorf("Only POST request allowed")
	}
	return nil
}

func (tgr *HTTP) handleFunc(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Read request body error\n"))
		return
	}

	if tgr.validateFunc != nil {
		err := tgr.validateFunc(r, body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Request validation error: %s\n", err)))
			return
		}
	}

	var req interface{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Request body parse error: %s\n", err)))
		return
	}

	reqMap, ok := req.(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Request body is not a map\n")))
		return
	}

	tgr.requestParam <- reqMap

	w.Write([]byte(fmt.Sprintf("Request received successfully\n")))
}

func (tgr *HTTP) createServerUnixSocket() *http.Server {
	if tgr.UnixSocket == "" {
		return nil
	}

	os.Remove(tgr.UnixSocket)

	ln, err := net.Listen("unix", tgr.UnixSocket)
	if err != nil {
		tgr.Logger.Errorf("[Trigger(HTTP)] Failed to listen to unix socket: %s", err)
		return nil
	}

	err = os.Chmod(tgr.UnixSocket, 0777)
	if err != nil {
		tgr.Logger.Errorf("[Trigger(HTTP)] Failed to chmod unix socket: %s", err)
		return nil
	}

	srv := &http.Server{
		Handler: http.HandlerFunc(tgr.handleFunc),
	}

	tgr.Logger.Infof("[Trigger(HTTP)] Starting server on unix socket: %s", tgr.UnixSocket)

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			tgr.Logger.Errorf("[Trigger(HTTP)] Server listen on unix socket error: %s", err)
		}
	}()

	return srv
}

func (tgr *HTTP) createServerTCPPort() *http.Server {
	if tgr.Port == 0 {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", tgr.Host, tgr.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		tgr.Logger.Errorf("[Trigger(HTTP)] Failed to listen to tcp port: %s", err)
		return nil
	}

	tgr.Logger.Infof("[Trigger(HTTP)] Starting server on tcp port: %s", addr)

	srv := &http.Server{
		Handler: http.HandlerFunc(tgr.handleFunc),
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			tgr.Logger.Errorf("[Trigger(HTTP)] Server listen on tcp port error: %s", err)
		}
	}()

	return srv
}

func (tgr *HTTP) run(ctx context.Context, param chan map[string]interface{}) {
	tgr.requestParam = make(chan map[string]interface{})

	srvUnix := tgr.createServerUnixSocket()
	srvTCP := tgr.createServerTCPPort()

	for {
		select {
		case <-ctx.Done():
			if srvUnix != nil {
				if err := srvUnix.Shutdown(context.Background()); err != nil {
					tgr.Logger.Errorf("[Trigger(HTTP)] HTTP server on unix socket shutdown error: %s", err)
				}
			}
			if srvTCP != nil {
				if err := srvTCP.Shutdown(context.Background()); err != nil {
					tgr.Logger.Errorf("[Trigger(HTTP)] HTTP server on tcp port shutdown error: %s", err)
				}
			}
			return
		case reqParam := <-tgr.requestParam:
			param <- reqParam
		}
	}
}

// Run the HTTP trigger
func (tgr *HTTP) Run(ctx context.Context, param chan map[string]interface{}) {
	tgr.validateFunc = tgr.validateMethod
	tgr.run(ctx, param)
}

// SetParam will set param from a map
func (tgr *HTTP) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&tgr.UnixSocket, param, "unix_socket")
	util.UpdateStringParam(&tgr.Host, param, "host")
	util.UpdateIntParam(&tgr.Port, param, "port")
}
