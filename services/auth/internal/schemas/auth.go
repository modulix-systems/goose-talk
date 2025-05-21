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
)
