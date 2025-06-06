package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOTP(t *testing.T) {
	securityProvider := SecurityProvider{otpLen: 6}
	otp := securityProvider.GenerateOTPCode()
	anotherOtp := securityProvider.GenerateOTPCode()
	assert.Equal(t, 6, len(otp))
	assert.NotEqual(t, otp, anotherOtp)
	const allowedChars = "1234567890"
	for _, otpChar := range otp {
		var isAllowedChar bool
		for _, allowedChar := range allowedChars {
			if otpChar == allowedChar {
				isAllowedChar = true
				break
			}
		}
		assert.Equal(t, true, isAllowedChar)
	}
}
