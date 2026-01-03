package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math"
)

func createRandBytes(size int) []byte {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

func (s *SecurityProvider) GenerateSessionId() string {
	return hex.EncodeToString(createRandBytes(8))
}

func (s *SecurityProvider) GeneratePrivateKey() string {
	return hex.EncodeToString(createRandBytes(32))
}

func (s *SecurityProvider) GenerateSecretTokenUrlSafe(len int) string {
	// base64 encodes every 3 bytes into a 4 characters
	bytesSize := int(math.Floor(float64(len) * 3 / 4)) 
	return base64.URLEncoding.EncodeToString(createRandBytes(bytesSize))
}
