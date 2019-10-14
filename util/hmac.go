package util

import (
	"crypto/hmac"
	"crypto/sha256"
)

// ValidateMAC checks the validation of message based on the key
func ValidateMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
