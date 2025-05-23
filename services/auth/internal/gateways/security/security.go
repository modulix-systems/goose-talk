package security

import (
	"crypto/rand"
	"encoding/base64"
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

func (s *SecurityProvider) GenerateSecretTokenUrlSafe(len int) string {
	tokenBytes := make([]byte, len*6/8)
	rand.Read(tokenBytes)
	return base64.URLEncoding.EncodeToString(tokenBytes)
}
