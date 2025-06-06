package gateways

import "errors"

var (
	ErrInvalidCredential = errors.New("invalid webauthn credential")
	ErrExpiredToken      = errors.New("expired token")
)
