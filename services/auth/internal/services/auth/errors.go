package auth

import "errors"

var (
	ErrInvalidSignUpCode = errors.New("Invalid signup code")
	ErrExpiredSignUpCode = errors.New(
		"Your signup code has expired! Please obtain a new one and try again",
	)
	ErrUserAlreadyExists  = errors.New("User with provided email already exists")
	ErrInvalidCredentials = errors.New("Invalid credentials")
)
