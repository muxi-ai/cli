package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// GenerateHMACSignature generates an HMAC-SHA256 signature for authentication
func GenerateHMACSignature(secretKey, method, path string) (string, int64) {
	timestamp := time.Now().Unix()

	// Strip query params from path for signature (use only path portion)
	signPath := path
	if idx := strings.Index(path, "?"); idx != -1 {
		signPath = path[:idx]
	}

	message := fmt.Sprintf("%d;%s;%s", timestamp, method, signPath)

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature, timestamp
}

// BuildAuthHeader builds the Authorization header for MUXI Server
func BuildAuthHeader(keyID, secretKey, method, path string) string {
	signature, timestamp := GenerateHMACSignature(secretKey, method, path)
	return fmt.Sprintf("MUXI-HMAC key=%s, timestamp=%d, signature=%s",
		keyID, timestamp, signature)
}
