package util

import (
	"crypto/rand"
	"encoding/base64"
)

func RandomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe token
	return base64.RawURLEncoding.EncodeToString(b), nil
}
