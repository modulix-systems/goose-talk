package gateways

import (
	"context"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"
)

type (
	NotificationsRepo interface {
	}
	MailClient interface {
		SendSignUpNotice(ctx context.Context, to string, data notifications.SignUpNotice, lang notifications.Language) error
		SendLoginNewDeviceNotice(ctx context.Context, to string, data notifications.LoginNewDeviceNotice, lang notifications.Language) error
		SendAccountDeactivatedNotice(ctx context.Context, to string, data notifications.AccountDeactivatedNotice, lang notifications.Language) error
		SendVerifyEmailNotice(ctx context.Context, to string, data notifications.EmailVerifyNotice, lang notifications.Language) error
		SendConfirmEmailTwoFaNotice(ctx context.Context, to string, data notifications.EmailTwoFaNotice, lang notifications.Language) error
		SendConfirmedTwoFaNotice(ctx context.Context, to string, data notifications.TwoFaConfirmedNotice, lang notifications.Language) error
	}
)
