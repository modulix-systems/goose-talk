package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (s *SecurityProvider) HashPassword(plainPassword string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func (s *SecurityProvider) ComparePasswords(hashed []byte, plain string) (error) {
	err := bcrypt.CompareHashAndPassword(hashed, []byte(plain))
	if err != nil {
		return fmt.Errorf("SecurityProvider - ComparePasswords - bcrypt.CompareHashAndPassword: %w", err)
	}
	return nil
}
