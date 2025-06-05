package gateways

import "errors"

var (
	ErrInvalidCredential = errors.New("Invalid webauthn credential")
	ErrExpiredToken      = errors.New("Expired token")
)
