package schemas

import "github.com/modulix-systems/goose-talk/internal/entity"

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
	}
	Verify2FASchema struct {
		TwoFATyp entity.TwoFADeliveryMethod
		Email    string
		Code     string
		// SignInConfToken must be present only if TOTP 2fa type is used
		SignInConfToken string
	}
)
