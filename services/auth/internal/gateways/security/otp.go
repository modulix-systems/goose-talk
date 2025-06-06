package security

func (s *SecurityProvider) GenerateOTPCode() string {
	const otpChars = "1234567890"
	buf := createRandBytes(s.otpLen)
	otpCharsLength := len(otpChars)
	for i := 0; i < s.otpLen; i++ {
		buf[i] = otpChars[int(buf[i])%otpCharsLength]
	}
	return string(buf)
}
