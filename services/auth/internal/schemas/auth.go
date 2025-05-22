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
	}
	SignInSchema struct {
		Login    string
		Password string
		// DeviceInfo can be user-agent for browser or for example some platform info
		// if request is coming from mobile app
		DeviceInfo string
		ClientIP   string
	}
	Verify2FASchema struct {
		DeviceInfo string
		ClientIP   string
		TwoFATyp   entity.TwoFADeliveryMethod
		Email      string
		Code       string
		// SignInConfToken must be present only if TOTP 2fa type is used
		SignInConfToken string
	}
	SessionUpdatePayload struct {
		AccessToken   string
		DeactivatedAt *time.Time
		LastSeenAt    time.Time
	}
	Add2FASchema struct {
		UserEmail string
		Typ       entity.TwoFADeliveryMethod
		Contact   string
	}
	Confirm2FASchema struct {
		UserEmail string
		Typ       entity.TwoFADeliveryMethod
		Contact   string
		// TotpSecret should be empty if Typ is not TOTP_APP
		TotpSecret string
		// ConfirmationCode is not required if Typ is TOTP_APP
		ConfirmationCode string
	}
)
