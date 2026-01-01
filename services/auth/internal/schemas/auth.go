package schemas

import (
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
)

type (
	SignUpSchema struct {
		Username         string
		Password         string
		Email            string
		FirstName        string
		LastName         string
		ConfirmationCode string
		LoginInfoSchema
	}
	SignInSchema struct {
		Login      string
		Password   string
		RememberMe bool
		LoginInfoSchema
	}
	LoginInfoSchema struct {
		IPAddr string
		// DeviceInfo can be user-agent for browser or for example some platform info
		// if request is coming from mobile app
		DeviceInfo string
	}
	Verify2FASchema struct {
		LoginInfoSchema
		TwoFATyp   entity.TwoFATransport
		Email      string
		Code       string
		RememberMe bool
		// SignInConfToken must be present only if TOTP 2fa type is used
		SignInConfToken string
	}
	SessionUpdatePayload struct {
		LastSeenAt time.Time
		ExpiresAt  time.Time
	}
	Add2FASchema struct {
		UserId  int
		Typ     entity.TwoFATransport
		Contact string
	}
	Confirm2FASchema struct {
		UserId  int
		Typ     entity.TwoFATransport
		Contact string
		// TotpSecret can be ommited if Typ is not TOTP_APP
		TotpSecret string
		// ConfirmationCode is not required if Typ is TOTP_APP
		ConfirmationCode string
	}
	ExportLoginTokenSchema struct {
		ClientId string
		LoginInfoSchema
	}
)
