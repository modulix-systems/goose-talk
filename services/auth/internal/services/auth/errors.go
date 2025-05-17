package auth

import "errors"

var (
	ErrInvalidSignUpCode = errors.New("Invalid signup code")
	ErrExpiredSignUpCode = errors.New(
		"Your signup code has expired! Please obtain a new one and try again",
	)
	ErrUserAlreadyExists    = errors.New("User with provided email already exists")
	ErrInvalidCredentials   = errors.New("Invalid credentials")
	ErrUnsupported2FAMethod = errors.New("Two factor authentication method associated with your account is not supported. Please use another one")
	ErrDisabledAccount      = errors.New("Your account is disabled. Try to contact support to resolve this issue")
)
