package auth

import "errors"

var (
	// ErrInvalidOtp          = errors.New("Invalid otp")
	ErrOTPInvalidOrExpired = errors.New("Entered code is invalid or expired. Please take a new one and try again")
	// ErrOtpExpired          = errors.New(
	// 	"Your otp has expired! Please obtain a new one and try again",
	// )
	// ErrInvalidOrExpiredTOTP = errors.New(
	// 	"Entered TOTP code is invalid or expired. Please take a new one from your auth application",
	// )
	Err2FANotEnabled = errors.New(
		"You have not enabled two factor authentication for your account",
	)
	ErrUserAlreadyExists    = errors.New("User with provided email already exists")
	ErrUserNotFound         = errors.New("User with provided email does not exist")
	ErrInvalidCredentials   = errors.New("Invalid credentials")
	ErrUnsupported2FAMethod = errors.New(
		"Two factor authentication method associated with your account is not supported. Please use another one",
	)
	Err2FaAlreadyAdded = errors.New("Two factor authentication is already associated with your account.")
	ErrDisabledAccount = errors.New(
		"Your account is disabled. Try to contact support to resolve this issue",
	)
	ErrSessionNotFound                  = errors.New("No active session found")
	ErrInvalidLoginToken                = errors.New("Your login token is invalid. Please obtain a new one")
	ErrExpiredLoginToken                = errors.New("Your login token has expired. Please obtain a new one")
	ErrInvalidPasskeyCredential         = errors.New("Invalid passkey credential")
	ErrPasskeyRegistrationNotInProgress = errors.New("Passkey registration is not in progress. Try to begin registration again")
)
