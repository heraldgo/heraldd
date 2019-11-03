package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

// HTTPServer is a trigger which will listen to http request
type HTTPServer struct {
	BaseLogger
	UnixSocket   string
	Host         string
	Port         int
	ServerHeader string
	ValidateFunc func(*http.Request, []byte) error
	ProcessFunc  func(http.ResponseWriter, *http.Request, []byte)

	srvUnix *http.Server
	srvTCP  *http.Server
}

func (h *HTTPServer) handleFunc(w http.ResponseWriter, r *http.Request) {
	if h.ServerHeader != "" {
		w.Header().Set("Server", h.ServerHeader)
	} else {
		w.Header().Set("Server", "herald")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Read request body error\n"))
		return
	}

	if h.ValidateFunc != nil {
		err := h.ValidateFunc(r, body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Request validation error: %s\n", err)))
			return
		}
	}

	if h.ProcessFunc == nil {
		w.Write([]byte(fmt.Sprintf("Request received successfully\n")))
		return
	}

	h.ProcessFunc(w, r, body)
}

func (h *HTTPServer) createServerUnixSocket() {
	if h.UnixSocket == "" {
		return
	}

	os.Remove(h.UnixSocket)

	ln, err := net.Listen("unix", h.UnixSocket)
	if err != nil {
		h.Errorf("Failed to listen to unix socket: %s", err)
		return
	}

	err = os.Chmod(h.UnixSocket, 0777)
	if err != nil {
		h.Errorf("Failed to chmod unix socket: %s", err)
		return
	}

	srv := &http.Server{
		Handler: http.HandlerFunc(h.handleFunc),
	}

	h.Infof("Starting server on unix socket: %s", h.UnixSocket)

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			h.Errorf("Server listen on unix socket error: %s", err)
		}
	}()

	h.srvUnix = srv
}

func (h *HTTPServer) createServerTCPPort() {
	if h.Port == 0 {
		return
	}

	addr := fmt.Sprintf("%s:%d", h.Host, h.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		h.Errorf("Failed to listen to tcp port: %s", err)
		return
	}

	h.Infof("Starting server on tcp port: %s", addr)

	srv := &http.Server{
		Handler: http.HandlerFunc(h.handleFunc),
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			h.Errorf("Server listen on tcp port error: %s", err)
		}
	}()

	h.srvTCP = srv
}

func (h *HTTPServer) shutdownServerUnixSocket() {
	if h.srvUnix != nil {
		if err := h.srvUnix.Shutdown(context.Background()); err != nil {
			h.Errorf("HTTPServer server on unix socket shutdown error: %s", err)
		}
		h.srvUnix = nil
	}
}

func (h *HTTPServer) shutdownServerTCPPort() {
	if h.srvTCP != nil {
		if err := h.srvTCP.Shutdown(context.Background()); err != nil {
			h.Errorf("HTTPServer server on tcp port shutdown error: %s", err)
		}
		h.srvTCP = nil
	}
}

// Start the http server
func (h *HTTPServer) Start() {
	h.createServerUnixSocket()
	h.createServerTCPPort()
}

// Stop the http server
func (h *HTTPServer) Stop() {
	h.shutdownServerUnixSocket()
	h.shutdownServerTCPPort()
}
