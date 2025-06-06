package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"time"
)

type SecurityProvider struct {
	totpTTL *time.Duration
	otpLen  int
}

func New(totpTTL *time.Duration, otpLen int) *SecurityProvider {
	return &SecurityProvider{
		totpTTL: totpTTL,
		otpLen:  otpLen,
	}
}

func createRandBytes(size int) []byte {
	buf := make([]byte, size)
	rand.Read(buf)
	return buf
}

func (s *SecurityProvider) GenerateSessionId() string {
	return hex.EncodeToString(createRandBytes(8))
}

func (s *SecurityProvider) GenerateSecretTokenUrlSafe(entropy int) string {
	return base64.URLEncoding.EncodeToString(createRandBytes(entropy))
}
