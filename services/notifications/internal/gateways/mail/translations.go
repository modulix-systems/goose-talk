package mailclient

import "github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"

type TranslationsTable = map[notifications.Language]map[notifications.EmailType]string

func getEmailSubject(typ notifications.EmailType, lang notifications.Language) string {
	translations := TranslationsTable{
		notifications.LANGUAGE_EN: {
			notifications.EMAIL_TYPE_VERIFY_EMAIL:        "Confirm your email",
			notifications.EMAIL_TYPE_SIGN_UP:             "Welcome! Complete your sign-up",
			notifications.EMAIL_TYPE_ACCOUNT_DEACTIVATED: "Your account has been deactivated",
			notifications.EMAIL_TYPE_LOGIN_NEW_DEVICE:    "New login from a new device",
			notifications.EMAIL_TYPE_EMAIL_TWO_FA:        "Your two-factor authentication code",
			notifications.EMAIL_TYPE_TWO_FA_CONFIRMED:    "Two-factor authentication enabled",
		},

		notifications.LANGUAGE_RU: {
			notifications.EMAIL_TYPE_VERIFY_EMAIL:        "Подтвердите электронную почту",
			notifications.EMAIL_TYPE_SIGN_UP:             "Добро пожаловать! Завершите регистрацию",
			notifications.EMAIL_TYPE_ACCOUNT_DEACTIVATED: "Ваш аккаунт был деактивирован",
			notifications.EMAIL_TYPE_LOGIN_NEW_DEVICE:    "Вход с нового устройства",
			notifications.EMAIL_TYPE_EMAIL_TWO_FA:        "Код двухфакторной аутентификации",
			notifications.EMAIL_TYPE_TWO_FA_CONFIRMED:    "Двухфакторная аутентификация включена",
		},
	}

	translation, ok := translations[lang]
	if !ok {
		translation = translations[notifications.LANGUAGE_EN]
	}

	return translation[typ]
}
