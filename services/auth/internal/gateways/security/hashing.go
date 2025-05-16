package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(plainPassword string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func ComparePasswords(hashed []byte, plain string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashed, []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
