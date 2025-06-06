package schemas

import (
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
)

type (
	SendConfirmationCodeSchema struct {
		email string
	}
	SignUpSchema struct {
		Username         string
		Password         string
		Email            string
		FirstName        string
		LastName         string
		ConfirmationCode string
		ClientIdentitySchema
	}
	SignInSchema struct {
		Login    string
		Password string
		ClientIdentitySchema
	}
	ClientIdentitySchema struct {
		IPAddr string
		// DeviceInfo can be user-agent for browser or for example some platform info
		// if request is coming from mobile app
		DeviceInfo string
	}
	Verify2FASchema struct {
		ClientIdentitySchema
		TwoFATyp entity.TwoFADeliveryMethod
		Email    string
		Code     string
		// SignInConfToken must be present only if TOTP 2fa type is used
		SignInConfToken string
	}
	SessionUpdatePayload struct {
		DeactivatedAt *time.Time
		LastSeenAt    time.Time
		ExpiresAt     time.Time
	}
	Add2FASchema struct {
		UserId  int
		Typ     entity.TwoFADeliveryMethod
		Contact string
	}
	Confirm2FASchema struct {
		UserId  int
		Typ     entity.TwoFADeliveryMethod
		Contact string
		// TotpSecret can be ommited if Typ is not TOTP_APP
		TotpSecret string
		// ConfirmationCode is not required if Typ is TOTP_APP
		ConfirmationCode string
	}
	ExportLoginTokenSchema struct {
		ClientId string
		ClientIdentitySchema
	}
)
