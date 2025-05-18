package security

import (
	"crypto/rand"
	"fmt"
)

func (s *SecurityProvider) GenerateOTPCode() string {
	buf := make([]byte, s.otpLen)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}
