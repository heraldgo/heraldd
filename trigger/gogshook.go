package trigger

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/xianghuzhao/heraldd/util"
)

// GogsHook is a trigger which will listen to http request
type GogsHook struct {
	HTTP
	Secret string
}

func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
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

	if !checkMAC(body, signature, key) {
		return fmt.Errorf("Signature validation Error")
	}
	return nil
}

// Run the GogsHook trigger
func (tgr *GogsHook) Run(ctx context.Context, param chan map[string]interface{}) {
	tgr.validateFunc = tgr.validateGogs
	tgr.run(ctx, param)
}

// SetParam will set param from a map
func (tgr *GogsHook) SetParam(param map[string]interface{}) {
	tgr.HTTP.SetParam(param)
	util.GetStringParam(&tgr.Secret, param, "secret")
}
