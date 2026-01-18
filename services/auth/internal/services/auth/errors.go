package auth

import "errors"

var (
	ErrOtpIsNotValid = errors.New("entered code is invalid or expired. Please obtain a new one and try again")
	Err2FANotEnabled = errors.New(
		"you have not enabled two factor authentication for your account",
	)
	ErrUserAlreadyExists    = errors.New("user with provided email already exists")
	ErrEmailUnverified      = errors.New("email verification is required to proceed")
	ErrUserNotFound         = errors.New("user with provided email does not exist")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUnsupported2FAMethod = errors.New(
		"two factor authentication method associated with your account is not supported. Please use another one",
	)
	Err2FaAlreadyAdded    = errors.New("two factor authentication is already associated with your account")
	ErrDeactivatedAccount = errors.New(
		"your account has been deactivated. Try to contact support to resolve this issue",
	)
	ErrSessionNotFound                  = errors.New("no active session found")
	ErrInvalidLoginToken                = errors.New("your login token is invalid. Please obtain a new one")
	ErrExpiredLoginToken                = errors.New("your login token has expired. Please obtain a new one")
	ErrInvalidPasskeyCredential         = errors.New("invalid passkey credential")
	ErrPasskeyRegistrationNotInProgress = errors.New("passkey registration is not in progress. Try to begin registration again")
)
