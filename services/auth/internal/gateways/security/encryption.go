package security

func (s *SecurityProvider) EncryptSymmetric(plaintext string, key string) ([]byte, error) {
	return []byte(plaintext), nil
}

func (s *SecurityProvider) DecryptSymmetric(encrypted []byte, key string) (string, error) {
	return string(encrypted), nil
}
