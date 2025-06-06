package auth

import "errors"

var (
	ErrOTPInvalidOrExpired = errors.New("entered code is invalid or expired. Please take a new one and try again")
	Err2FANotEnabled       = errors.New(
		"you have not enabled two factor authentication for your account",
	)
	ErrUserAlreadyExists    = errors.New("user with provided email already exists")
	ErrUserNotFound         = errors.New("user with provided email does not exist")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUnsupported2FAMethod = errors.New(
		"two factor authentication method associated with your account is not supported. Please use another one",
	)
	Err2FaAlreadyAdded = errors.New("two factor authentication is already associated with your account")
	ErrDisabledAccount = errors.New(
		"your account is disabled. Try to contact support to resolve this issue",
	)
	ErrSessionNotFound                  = errors.New("no active session found")
	ErrInvalidLoginToken                = errors.New("your login token is invalid. Please obtain a new one")
	ErrExpiredLoginToken                = errors.New("your login token has expired. Please obtain a new one")
	ErrInvalidPasskeyCredential         = errors.New("invalid passkey credential")
	ErrPasskeyRegistrationNotInProgress = errors.New("passkey registration is not in progress. Try to begin registration again")
)
