package dtos

import "github.com/modulix-systems/goose-talk/internal/entity"

type (
	Verify2FARequest struct {
		TwoFATyp   entity.TwoFaMethod
		Email      string
		Code       string
		RememberMe bool
		IpAddr     string
		DeviceInfo string
		// SignInConfirmationCode must be present only if TOTP 2fa type is used
		SignInConfirmationCode string
	}
	Add2FARequest struct {
		UserId  int
		Typ     entity.TwoFaMethod
		Contact string
	}
	Confirm2FARequest struct {
		UserId  int
		Typ     entity.TwoFaMethod
		Contact string
		// TotpSecret can be ommited if Typ is not TOTP_APP
		TotpSecret string
		// ConfirmationCode is not required if Typ is TOTP_APP
		ConfirmationCode string
	}
)
