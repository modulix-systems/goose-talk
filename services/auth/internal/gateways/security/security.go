package security

import (
	"time"
)

type SecurityProvider struct {
	totpTTL time.Duration
	otpLen  int
	appName string
}

func New(totpTTL time.Duration, otpLen int, appName string) *SecurityProvider {
	return &SecurityProvider{
		totpTTL: totpTTL,
		otpLen:  otpLen,
		appName: appName,
	}
}
