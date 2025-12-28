package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
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

func (s *SecurityProvider) GenerateSecretTokenUrlSafe(entropy int) string {
	return base64.URLEncoding.EncodeToString(createRandBytes(entropy))
}
