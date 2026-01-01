package security

func (s *SecurityProvider) GenerateOTPCode() string {
	const characters = "1234567890"
	buf := createRandBytes(s.otpLen)
	otpCharsLength := len(characters)
	for i := 0; i < s.otpLen; i++ {
		buf[i] = characters[int(buf[i])%otpCharsLength]
	}
	return string(buf)
}
