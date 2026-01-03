package dtos

import "github.com/modulix-systems/goose-talk/internal/entity"

type SignInRequest struct {
	Login      string `validate:"required"`
	Password   string `validate:"required,min=8"`
	RememberMe bool
	IpAddr     string `validate:"required,ip"`
	DeviceInfo string `validate:"required"`
}

type SignInResponse struct {
	// ConfirmationCode is present if user has totp 2fa verification method
	// it should be used in Verify2FA to prove that user has went through signin first before trying to verify 2fa
	ConfirmationCode string
	User             *entity.User
	Session          *entity.AuthSession
}
