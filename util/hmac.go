package util

import (
	"crypto/hmac"
	"crypto/sha256"
)

// CalculateMAC calculate the HMAC
func CalculateMAC(message, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}

// ValidateMAC checks the validation of message based on the key
func ValidateMAC(message, messageMAC, key []byte) bool {
	expectedMAC := CalculateMAC(message, key)
	return hmac.Equal(messageMAC, expectedMAC)
}
