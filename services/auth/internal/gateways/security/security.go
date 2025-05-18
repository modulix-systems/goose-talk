package security

import "time"

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
