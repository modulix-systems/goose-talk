package mail

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"
)

func (s *Service) SendMail(ctx context.Context, email notifications.EmailMessage) error {
	switch email.Type {
	case notifications.EMAIL_TYPE_SIGN_UP:
		var data notifications.SignUpNotice
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return fmt.Errorf("mail - Service.SendMail - signup - json.Unmarshal: %w", err)
		}
		return s.mailClient.SendSignUpNotice(ctx, email.To, data, email.Language)

	case notifications.EMAIL_TYPE_VERIFY_EMAIL:
		var data notifications.EmailVerifyNotice
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return fmt.Errorf("mail - Service.SendMail - verify email - json.Unmarshal: %w", err)
		}
		return s.mailClient.SendVerifyEmailNotice(ctx, email.To, data, email.Language)

	case notifications.EMAIL_TYPE_ACCOUNT_DEACTIVATED:
		var data notifications.AccountDeactivatedNotice
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return fmt.Errorf("mail - Service.SendMail - account deactivated - json.Unmarshal: %w", err)
		}
		return s.mailClient.SendAccountDeactivatedNotice(ctx, email.To, data, email.Language)

	case notifications.EMAIL_TYPE_LOGIN_NEW_DEVICE:
		var data notifications.LoginNewDeviceNotice
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return fmt.Errorf("mail - Service.SendMail - login new device - json.Unmarshal: %w", err)
		}
		return s.mailClient.SendLoginNewDeviceNotice(ctx, email.To, data, email.Language)

	case notifications.EMAIL_TYPE_EMAIL_TWO_FA:
		var data notifications.EmailTwoFaNotice
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return fmt.Errorf("mail - Service.SendMail - email two fa - json.Unmarshal: %w", err)
		}
		return s.mailClient.SendConfirmEmailTwoFaNotice(ctx, email.To, data, email.Language)

	case notifications.EMAIL_TYPE_TWO_FA_CONFIRMED:
		var data notifications.TwoFaConfirmedNotice
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return fmt.Errorf("mail - Service.SendMail - two fa confirmed - json.Unmarshal: %w", err)
		}
		return s.mailClient.SendConfirmedTwoFaNotice(ctx, email.To, data, email.Language)
	}

	s.log.Error("mail - service.SendMail - unknown email type", "type", email.Type)
	return fmt.Errorf("mail - service.SendMail - unknown email type: %s", email.Type)
}
