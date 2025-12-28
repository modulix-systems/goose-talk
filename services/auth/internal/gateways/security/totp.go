package security

func (s *SecurityProvider) GenerateTOTPEnrollUrlWithSecret(accName string) (string, string) {
	return "", ""
}

func (s *SecurityProvider) ValidateTOTP(code string, secret string) bool {
	return true
}
